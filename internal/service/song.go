package service

import (
	"awesomeProject/internal/model"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

type SongServiceInterface interface {
	GetSongs(filterID int64, filterById bool, filterGroup, filterSong string, page, pageSize int) (model.SongsResponse, error)
	AddSong(song model.Song) (model.Song, error)
	DeleteSong(id int64) error
	UpdateSong(id int64, updateSong model.Song) (model.Song, error)
	GetSongVerses(id int64, versePage, verseSize int) (model.VerseResponse, error)
	SearchVerses(searchText string) ([]model.SongVerse, error)
}

type SongService struct {
	db     *sql.DB
	logger *slog.Logger // Используем *slog.Logger
}

func NewSongService(db *sql.DB, logger *slog.Logger) *SongService {
	return &SongService{db: db, logger: logger}
}

type SongVerse struct {
	SongID int64  `json:"song_id"`
	Group  string `json:"group"`
	Song   string `json:"song"`
	Verse  string `json:"verse"`
}

func (s *SongService) SearchVerses(searchText string) ([]model.SongVerse, error) {
	s.logger.Debug("Searching verses", "text", searchText)

	query := `
	SELECT id, "group", song, text FROM songs WHERE text ILIKE $1
	`
	rows, err := s.db.Query(query, "%"+searchText+"%")
	if err != nil {
		s.logger.Error("Failed to search verses", "error", err)
	}
	defer rows.Close()

	var results []model.SongVerse
	for rows.Next() {
		var song model.Song
		if err := rows.Scan(&song.ID, &song.Group, &song.Song, &song.Text); err != nil {
			s.logger.Error("Failed to scan song row", "error", err)
			return nil, fmt.Errorf("ошибка чтения данных: %v", err)
		}
		verses := strings.Split(song.Text, "\n")
		for _, verse := range verses {
			if strings.Contains(strings.ToLower(verse), strings.ToLower(searchText)) {
				results = append(results, model.SongVerse{
					SongID: song.ID,
					Group:  song.Group,
					Song:   song.Song,
					Verse:  verse,
				})
			}
		}
	}
	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating song rows", "error", err)
		return nil, fmt.Errorf("ошибка чтения строк: %v", err)
	}

	if len(results) == 0 {
		s.logger.Warn("No verses found", "text", searchText)
		return nil, fmt.Errorf("куплеты с текстом %q не найдены", searchText)
	}

	s.logger.Info("Verses found", "count", len(results))
	return results, nil
}
func (s *SongService) GetSongs(filterID int64, filterById bool, filterGroup, filterSong string, page, pageSize int) (model.SongsResponse, error) {
	s.logger.Debug("Fetching songs", "filter_id", filterID, "filter_group", filterGroup, "filter_song", filterSong)
	query := `SELECT id, "group", song, text FROM songs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM songs WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filterById {
		query += fmt.Sprintf(" AND id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND id = $%d", argIndex)
		args = append(args, filterID)
		argIndex++
	}
	if filterGroup != "" {
		query += fmt.Sprintf(" AND LOWER(\"group\") LIKE $%d", argIndex)
		countQuery += fmt.Sprintf(" AND LOWER(\"group\") LIKE $%d", argIndex)
		args = append(args, "%"+filterGroup+"%")
		argIndex++
	}
	if filterSong != "" {
		query += fmt.Sprintf(" AND LOWER(\"song\") LIKE $%d", argIndex)
		countQuery += fmt.Sprintf(" AND LOWER(\"song\") LIKE $%d", argIndex)
		args = append(args, "%"+filterSong+"%")
		argIndex++
	}

	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		s.logger.Error("Failed to count songs", "error", err)
		return model.SongsResponse{}, fmt.Errorf("ошибка подсчёта записей: %v", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		s.logger.Warn("Requested page exceeds total pages", "page", page, "total_pages", totalPages)
		return model.SongsResponse{}, fmt.Errorf("запрошенная страница превышает количество страниц")
	}

	offset := (page - 1) * pageSize
	query += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.logger.Error("Failed to query songs", "error", err)
		return model.SongsResponse{}, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var songs []model.Song
	for rows.Next() {
		var song model.Song
		if err := rows.Scan(&song.ID, &song.Group, &song.Song, &song.Text); err != nil {
			s.logger.Error("Failed to scan song row", "error", err)
			return model.SongsResponse{}, fmt.Errorf("ошибка чтения данных: %v", err)
		}
		songs = append(songs, song)
	}
	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating song rows", "error", err)
		return model.SongsResponse{}, fmt.Errorf("ошибка чтения строк: %v", err)
	}

	s.logger.Info("Songs fetched successfully", "count", len(songs))
	return model.SongsResponse{
		Items:      songs,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *SongService) AddSong(song model.Song) (model.Song, error) {
	err := s.db.QueryRow(
		`INSERT INTO songs ("group", song) VALUES ($1, $2) RETURNING id`,
		song.Group, song.Song,
	).Scan(&song.ID)
	if err != nil {
		return model.Song{}, fmt.Errorf("ошибка добавления песни: %v", err)
	}
	return song, nil
}
func (s *SongService) DeleteSong(id int64) error {
	result, err := s.db.Exec(`DELETE FROM songs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления песни: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки результата: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("песня не найдена")
	}
	return nil
}

func (s *SongService) UpdateSong(id int64, updateSong model.Song) (model.Song, error) {
	query := `UPDATE songs SET `
	var args []interface{}
	argIndex := 1

	if updateSong.Group != "" {
		query += fmt.Sprintf(`"group" = $%d, `, argIndex)
		args = append(args, updateSong.Group)
		argIndex++
	}
	if updateSong.Song != "" {
		query += fmt.Sprintf(`song = $%d, `, argIndex)
		args = append(args, updateSong.Song)
		argIndex++
	}
	if len(args) == 0 {
		return model.Song{}, fmt.Errorf("не указаны поля для обновления")
	}
	query = query[:len(query)-2]
	query += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, id)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return model.Song{}, fmt.Errorf("ошибка обновления песни: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Song{}, fmt.Errorf("ошибка проверки результата: %v", err)
	}
	if rowsAffected == 0 {
		return model.Song{}, fmt.Errorf("песня с заданным ID не найдена")
	}
	var updatedSong model.Song
	err = s.db.QueryRow(`SELECT id, "group", song FROM songs WHERE id = $1`, id).Scan(&updatedSong.ID, &updatedSong.Group, &updatedSong.Song)
	if err != nil {
		return model.Song{}, fmt.Errorf("ошибка получения новой песни: %v", err)
	}
	return updatedSong, nil
}

func (s *SongService) GetSongVerses(id int64, versePage, verseSize int) (model.VerseResponse, error) {
	var text string
	err := s.db.QueryRow(`SELECT text FROM songs WHERE id = $1`, id).Scan(&text)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.VerseResponse{}, fmt.Errorf("песня с ID %d не найдена", id)
		}
		return model.VerseResponse{}, fmt.Errorf("ошибка получения текста песни: %v", err)
	}

	//Разделение текста на куплеты
	verses := strings.Split(text, "\n")
	totalVerses := len(verses)

	totalPages := (totalVerses + verseSize - 1) / verseSize
	if totalPages == 0 {
		totalPages = 1
	}

	if versePage > totalPages {
		return model.VerseResponse{}, fmt.Errorf("запрошенная страница куплетов превышает общее кол-во страниц")
	}

	start := (versePage - 1) * verseSize
	end := start + verseSize
	if end > totalVerses {
		end = totalVerses
	}

	pagedVerses := verses[start:end]

	return model.VerseResponse{
		Verses:      pagedVerses,
		VersePage:   versePage,
		VerseSize:   verseSize,
		TotalVerses: totalVerses,
		TotalPages:  totalPages,
	}, nil
}
