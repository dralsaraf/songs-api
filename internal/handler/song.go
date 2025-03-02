package handler

import (
	"awesomeProject/internal/model"
	"awesomeProject/internal/service"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service service.SongServiceInterface
	logger  *slog.Logger
}

func NewHandler(service service.SongServiceInterface, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func parseID(c echo.Context) (int64, error) {
	idStr := c.QueryParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("укажите корректный ID песни")
	}
	return id, nil
}
func (h *Handler) errorResponse(c echo.Context, status int, message string) error {
	if h.logger == nil {
		// Используем slog.Default() как fallback, если logger не инициализирован
		slog.Default().Error("Request failed, logger is nil", "method", c.Request().Method, "path", c.Request().URL.Path, "status", status, "message", message)
	} else {
		h.logger.Error("Request failed", "method", c.Request().Method, "path", c.Request().URL.Path, "status", status, "message", message)
	}
	return c.JSON(status, model.Response{
		Status:  "Error",
		Message: message,
	})
}

// GetHandler возвращает список песен с фильтрацией и пагинацией
// @Summary Получить список песен
// @Description Возвращает список песен с возможностью фильтрации по ID, группе, названию и пагинацией
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int false "ID песни"
// @Param group query string false "Название группы"
// @Param song query string false "Название песни"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(10)
// @Success 200 {object} model.SongsResponse
// @Failure 400 {object} model.Response "Неверные параметры запроса"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs [get]
func (h *Handler) GetHandler(c echo.Context) error {
	filterGroup := strings.ToLower(c.QueryParam("group"))
	filterSong := strings.ToLower(c.QueryParam("song"))
	filterIDStr := c.QueryParam("id")
	var filterID int64
	var filterByID bool

	if filterIDStr != "" {
		id, err := strconv.ParseInt(filterIDStr, 10, 64)
		if err != nil {
			return h.errorResponse(c, http.StatusBadRequest, "Неверный формат ID")
		}
		filterID = id
		filterByID = true
	}

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	resp, err := h.service.GetSongs(filterID, filterByID, filterGroup, filterSong, page, pageSize)
	if err != nil {
		if strings.Contains(err.Error(), "запрошенная страница") {
			return h.errorResponse(c, http.StatusBadRequest, err.Error())
		}
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

// PostHandler добавляет новую песню
// @Summary Добавить новую песню
// @Description Добавляет новую песню с указанными данными
// @Tags songs
// @Accept json
// @Produce json
// @Param song body model.Song true "Данные песни"
// @Success 200 {object} model.Response "Песня успешно добавлена"
// @Failure 400 {object} model.Response "Неверный формат данных или пустые поля"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs [post]
func (h *Handler) PostHandler(c echo.Context) error {
	var song model.Song
	if err := c.Bind(&song); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "Не смогли добавить песню: "+err.Error())
	}

	if song.Group == "" || song.Song == "" {
		return h.errorResponse(c, http.StatusBadRequest, "Группа или название песни не могут быть пустыми")
	}

	newSong, err := h.service.AddSong(song)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.Response{
		Status:  "Success",
		Message: "Песня добавлена",
		Data:    newSong,
	})
}

// DeleteHandler удаляет песню по ID
// @Summary Удалить песню
// @Description Удаляет песню по указанному ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "ID песни"
// @Success 200 {object} model.Response "Песня успешно удалена"
// @Failure 400 {object} model.Response "Неверный формат ID или песня не найдена"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs [delete]
func (h *Handler) DeleteHandler(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return h.errorResponse(c, http.StatusBadRequest, err.Error())
	}

	if err := h.service.DeleteSong(id); err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			return h.errorResponse(c, http.StatusBadRequest, err.Error())
		}
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.Response{
		Status:  "Success",
		Message: "Песня удалена",
	})
}

// PatchHandler обновляет данные песни
// @Summary Обновить песню
// @Description Обновляет данные песни по указанному ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "ID песни"
// @Param song body model.Song true "Обновляемые данные песни"
// @Success 200 {object} model.Response "Песня успешно обновлена"
// @Failure 400 {object} model.Response "Неверный формат данных или ID"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs [patch]
func (h *Handler) PatchHandler(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return h.errorResponse(c, http.StatusBadRequest, err.Error())
	}

	var updateSong model.Song
	if err := c.Bind(&updateSong); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
	}

	updatedSong, err := h.service.UpdateSong(id, updateSong)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") || strings.Contains(err.Error(), "не указаны поля") {
			return h.errorResponse(c, http.StatusBadRequest, err.Error())
		}
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.Response{
		Status:  "Success",
		Message: "Песня обновлена",
		Data:    updatedSong,
	})
}

// GetVersesHandler возвращает текст песни с пагинацией по куплетам
// @Summary Получить куплеты песни
// @Description Возвращает текст песни с пагинацией по куплетам по указанному ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param verse_page query int false "Номер страницы куплетов" default(1)
// @Param verse_size query int false "Размер страницы куплетов" default(1)
// @Success 200 {object} model.Response "Куплеты успешно получены"
// @Failure 400 {object} model.Response "Неверный формат ID или страницы"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs/{id}/verses [get]
func (h *Handler) GetVersesHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return h.errorResponse(c, http.StatusBadRequest, "укажите корректный ID песни")
	}

	versePage, err := strconv.Atoi(c.QueryParam("verse_page"))
	if err != nil || versePage < 1 {
		versePage = 1
	}

	verseSize, err := strconv.Atoi(c.QueryParam("verse_size"))
	if err != nil || verseSize < 1 {
		verseSize = 1 // Один куплет на одну страницу по дефолту
	}

	resp, err := h.service.GetSongVerses(id, versePage, verseSize)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") || strings.Contains(err.Error(), "превышает") {
			return h.errorResponse(c, http.StatusBadRequest, err.Error())
		}
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.Response{
		Status:  "Success",
		Message: "Куплеты получены",
		Data:    resp,
	})
}

// SearchVersesHandler ищет куплеты по тексту
// @Summary Поиск куплетов по тексту
// @Description Ищет куплеты, содержащие указанный текст
// @Tags songs
// @Accept json
// @Produce json
// @Param text query string true "Текст для поиска"
// @Success 200 {object} model.Response "Куплеты успешно найдены"
// @Failure 400 {object} model.Response "Текст для поиска не указан"
// @Failure 404 {object} model.Response "Куплеты не найдены"
// @Failure 500 {object} model.Response "Внутренняя ошибка сервера"
// @Router /songs/verses/search [get]
func (h *Handler) SearchVersesHandler(c echo.Context) error {
	searchText := c.QueryParam("text")
	if searchText == "" {
		return h.errorResponse(c, http.StatusBadRequest, "укажите текст для поиска")
	}

	h.logger.Info("Handing GET /songs/verses/search", "text", searchText)
	results, err := h.service.SearchVerses(searchText)
	if err != nil {
		if strings.Contains(err.Error(), "не найдены") {
			return h.errorResponse(c, http.StatusNotFound, err.Error())
		}
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	h.logger.Debug("Verses search completed", "count", len(results))
	return c.JSON(http.StatusOK, model.Response{
		Status:  "Success",
		Message: "Куплеты найдены",
		Data:    results,
	})
}
