package models

import "time"

// Song represents structure of song in library
type Song struct {
	ID          int       `json:"id"`
	Group       string    `json:"group"`
	Title       string    `json:"song"`
	ReleaseDate time.Time `json:"release_date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SongFilters holds optional fields to filter songs
type SongFilters struct {
	Group       string    `json:"group"`
	Title       string    `json:"song"`
	ReleaseDate time.Time `json:"release_date"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
