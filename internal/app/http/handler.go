package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"rest-songs/internal/app/api"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/postgresql"
)

// Handler struct wraps service interface, which interacts with business logic
type Handler struct {
	service     api.Service
	externalAPI string
	logger      *logrus.Logger
}

// New creates new Handler instance and takes api.Service and logger as parameters
func New(service api.Service, externalAPI string, logger *logrus.Logger) *Handler {
	return &Handler{
		service:     service,
		externalAPI: externalAPI,
		logger:      logger,
	}
}

// GetSongsHandler handles GET request for filtering and retrieving songs
// @Summary Get songs
// @Description Get songs with optional filters and pagination
// @Tags Songs
// @Accept json
// @Produce json
// @Param group query string false "Filter by group"
// @Param song query string false "Filter by song title"
// @Param release_date query string false "Filter by release date" Format("02.01.2006")
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {array} models.Song
// @Failure 400 {string} string "Неправильный формат данных"
// @Failure 500 {string} string "Проблема на сервере"
// @Router /songs [get]
func (h *Handler) GetSongsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse filter parameters
	group := query.Get("group")
	title := query.Get("song")
	releaseDateStr := query.Get("release_date")

	var releaseDate time.Time
	var err error
	if releaseDateStr != "" {
		// Parse release date from string to time.Time format
		releaseDate, err = time.Parse("02.01.2006", releaseDateStr)
		if err != nil {
			http.Error(w, "Неправильный формат даты", http.StatusBadRequest)
			return
		}
	}

	filter := models.SongFilters{
		Group:       group,
		Title:       title,
		ReleaseDate: releaseDate,
	}

	// Parse pagination parameters
	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	// Convert page string to integer
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1 // default to page 1
	}

	// Convert page size string to integer
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10 // default to 10 items per page
	}

	// Call service to get songs with filter
	songs, err := h.service.GetSongsWithFilter(filter, page, pageSize)
	if err != nil {
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with list of songs
	w.Header().Set("Content-Type", "application/json")
	if len(songs) == 0 {
		json.NewEncoder(w).Encode("Песня не найдена/Список песен пуст")
		return
	}
	json.NewEncoder(w).Encode(songs)
}

// GetSongTextHandler handles GET requests to retrieve paginated song text by song ID
// @Summary Get paginated song text
// @Description Get the verses of a song by its ID with optional pagination parameters
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of verses per page" default(10)
// @Success 200 {array} string "Array of verses"
// @Failure 400 {string} string "Неправильный формат ID или Страница выходит за пределы доступного диапазона"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Проблема на сервере"
// @Router /songs/text/{id} [get]
func (h *Handler) GetSongTextHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert ID string to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный формат ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters from query
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	// Convert page string to integer
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1 // default to page 1
	}

	// Convert page size string to integer
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10 // default to 10 items per page
	}

	// Call service to get paginated song text
	verses, err := h.service.GetSongText(id, page, pageSize)
	if err != nil {
		// Return 404 error if song not found
		if errors.Is(err, postgresql.ErrSongNotFound) {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
			return
		}

		// Return 400 error if pagination error occurs
		if errors.Is(err, api.ErrPageOutOfBounds) {
			http.Error(w, "Страница выходит за пределы доступного диапазона", http.StatusBadRequest)
			return
		}

		// Return 500 error for other issue
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with paginated verses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verses)
}

// UpdateSongByIdHandler handles PUT requests to update a song by its ID
// @Summary Update song by ID
// @Description Update an existing song's details by its ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body object true "Song data to update"
// @Success 200 {object} models.Song "Updated song object"
// @Failure 400 {string} string "Неправильный формат ID, Неправильный формат данных, or Неправильный формат даты"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Проблема на сервере"
// @Router /songs/{id} [put]
func (h *Handler) UpdateSongByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert ID string to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный формат ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Group       string `json:"group"`
		Title       string `json:"song"`
		ReleaseDate string `json:"release_date"`
		Text        string `json:"text"`
		Link        string `json:"link"`
	}

	// Decode request body into input struct
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	// Parse release date from string to time.Time format
	releaseDate, err := time.Parse("02.01.2006", input.ReleaseDate)
	if err != nil {
		http.Error(w, "Неправильный формат даты", http.StatusBadRequest)
		return
	}

	// Create new song object with updated data
	song := models.Song{
		Group:       input.Group,
		Title:       input.Title,
		ReleaseDate: releaseDate,
		Text:        input.Text,
		Link:        input.Link,
	}

	// Call service to update song by ID
	updatedSong, err := h.service.UpdateSongById(id, song)
	if err != nil {
		// Return 404 error if song not found
		if errors.Is(err, postgresql.ErrSongNotFound) {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
			return
		}

		// Return 500 error for other issue
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with updated song
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSong)
}

// DeleteSongByIdHandler handles DELETE requests to remove a song by its ID
// @Summary Delete song by ID
// @Description Delete an existing song by its ID from the database
// @Tags Songs
// @Param id path int true "Song ID"
// @Success 204 "No Content - Successfully deleted"
// @Failure 400 {string} string "Неправильный формат ID"
// @Failure 404 {string} string "Песня не найдена"
// @Failure 500 {string} string "Проблема на сервере"
// @Router /songs/{id} [delete]
func (h *Handler) DeleteSongByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert ID string to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный формат ID", http.StatusBadRequest)
		return
	}

	// Call service to delete song by ID
	err = h.service.DeleteSongById(id)
	if err != nil {
		// Return 404 error if song not found
		if errors.Is(err, postgresql.ErrSongNotFound) {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
			return
		}
		// Return 500 error for other issue
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with 204 status indicating successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// AddSongHandler handles POST requests to add a new song
// @Summary Add a new song
// @Description Create a new song by providing the group and song title
// @Tags Songs
// @Accept json
// @Produce json
// @Param song body models.AddSongRequest true "Song details"
// @Success 201 {object} models.Song "Created song"
// @Failure 400 {string} string "Неправильный формат данных"
// @Failure 500 {string} string "Проблема на сервере"
// @Router /songs [post]
func (h *Handler) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	var input models.AddSongRequest

	// Decode request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	h.logger.Infof("AddSongHandler[handler]: Получение деталей песни через API для группы: %s, песни: %s",
		input.Group, input.Song)

	// Encode group and song parameters for URL
	group := url.QueryEscape(input.Group)
	song := url.QueryEscape(input.Song)

	mockserverURL := h.externalAPI + "/info?group=" + group + "&song=" + song

	h.logger.Infof("AddSongHandler[handler]: mockserverURL: %s",
		mockserverURL)

	// get song details from mockserver
	resp, err := http.Get(mockserverURL)
	if err != nil {
		h.logger.Errorf("AddSongHandler[handler]: Ошибка отправки запроса к API: %v", err)
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("AddSongHandler[handler]: API вернул ошибку: %s", resp.Status)
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Parse API response
	var songDetails models.SongDetail
	if err = json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		h.logger.Errorf("AddSongHandler[handler]: Ошибка парсинга ответа от API: %v", err)
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("AddSongHandler[handler]: Успешно получены детали песни через API")

	// Call service to create song
	createdSong, err := h.service.CreateSong(input.Group, input.Song, songDetails)
	if err != nil {
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with created song
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSong)
}

// RegisterRoutes registers HTTP routes for song operations
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// API Routes
	// @Router /songs [get]
	r.HandleFunc("/songs", h.GetSongsHandler).Methods("GET")

	// @Router /songs/text/{id} [get]
	r.HandleFunc("/songs/text/{id}", h.GetSongTextHandler).Methods("GET")

	// @Router /songs/{id} [put]
	r.HandleFunc("/songs/{id}", h.UpdateSongByIdHandler).Methods("PUT")

	// @Router /songs/{id} [delete]
	r.HandleFunc("/songs/{id}", h.DeleteSongByIdHandler).Methods("DELETE")

	// @Router /songs [post]
	r.HandleFunc("/songs", h.AddSongHandler).Methods("POST")

	// Swagger documentation endpoint
	r.PathPrefix("/docs/swagger/").Handler(httpSwagger.WrapHandler)
	r.HandleFunc("/docs/swagger/index.html", httpSwagger.WrapHandler)
}
