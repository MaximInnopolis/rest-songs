package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"rest-songs/internal/app/api"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/postgresql"
)

// Handler struct wraps service interface, which interacts with business logic
type Handler struct {
	service api.Service
	logger  *logrus.Logger
}

// New creates new Handler instance and takes api.Service and logger as parameters
func New(service api.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

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

	// Respond with list of song
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

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

// UpdateSongByIdHandler handles HTTP PUT request to update existing song by ID
// It parses song ID and input data, calls service to update song, and returns updated song
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

// DeleteSongByIdHandler handles HTTP DELETE request to delete song by ID
// It parses song ID, calls service to delete song, and returns appropriate status
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

func (h *Handler) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Group string `json:"group"`
		Song  string `json:"song"`
	}

	// Decode request body into input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неправильный формат данных", http.StatusBadRequest)
		return
	}

	// Call service to create song
	createdSong, err := h.service.CreateSong(input.Group, input.Song)
	if err != nil {
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with created task
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSong)
}

// RegisterRoutes registers HTTP routes for song operations
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/songs", h.GetSongsHandler).Methods("GET")
	r.HandleFunc("/songs/{id}/text", h.GetSongTextHandler).Methods("GET")
	r.HandleFunc("/songs/{id}", h.UpdateSongByIdHandler).Methods("PUT")
	r.HandleFunc("/songs/{id}", h.DeleteSongByIdHandler).Methods("DELETE")
	r.HandleFunc("/songs", h.AddSongHandler).Methods("POST")
}
