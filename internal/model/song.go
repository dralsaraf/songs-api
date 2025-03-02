package model

type Song struct {
	ID    int64  `json:"ID"`
	Group string `json:"group"`
	Song  string `json:"song"`
	Text  string `json:"text,omitempty"`
}
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SongsResponse struct {
	Items      []Song `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
}

type VerseResponse struct {
	Verses      []string `json:"verses"`
	VersePage   int      `json:"verse_page"`
	VerseSize   int      `json:"verse_size"`
	TotalVerses int      `json:"total_verses"`
	TotalPages  int      `json:"total_pages"`
}

type SongVerse struct {
	SongID int64  `json:"song_id"`
	Group  string `json:"group"`
	Song   string `json:"song"`
	Verse  string `json:"verse"`
}
