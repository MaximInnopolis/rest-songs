package models

import "time"

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
