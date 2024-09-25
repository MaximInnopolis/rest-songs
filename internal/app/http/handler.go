package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"rest-songs/internal/app/api"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/postgresql"
)

// Handler struct wraps service interface, which interacts with business logic
type Handler struct {
	service api.Service
}

// New creates new Handler instance and takes api.Service as parameter
func New(service api.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSongsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	group := query.Get("group")
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
	releaseDate, err := time.Parse(time.RFC3339, input.ReleaseDate)
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
		// Return 404 error if song is not found
		if errors.Is(err, postgresql.ErrSongNotFound) {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
			return
		}

		// Return 500 error for any other issue
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
		// Return 404 error if song is not found
		if errors.Is(err, postgresql.ErrSongNotFound) {
			http.Error(w, "Песня не найдена", http.StatusNotFound)
			return
		}
		// Return 500 error for any other issue
		http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
		return
	}

	// Respond with 204 status indicating successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers HTTP routes for song operations
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/songs", h.GetSongsHandler).Methods("GET")
	r.HandleFunc("/songs/{id}", h.UpdateSongByIdHandler).Methods("PUT")
	r.HandleFunc("/songs/{id}", h.DeleteSongByIdHandler).Methods("DELETE")
}
