package api

import (
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/postgresql"
)

type Service interface {
	UpdateSongById(id int, song models.Song) (models.Song, error)
	DeleteSongById(id int) error
}

type SongService struct {
	repo postgresql.Repository
}

func New(repo postgresql.Repository) *SongService {
	return &SongService{repo: repo}
}

// UpdateSongById updates an existing song by ID using repository
// and returns updated song
func (s *SongService) UpdateSongById(id int, song models.Song) (models.Song, error) {
	return s.repo.Update(id, song)
}

// DeleteSongById deletes song by ID using repository
func (s *SongService) DeleteSongById(id int) error {
	return s.repo.Delete(id)
}
