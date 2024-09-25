package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/database"
)

var ErrSongNotFound = errors.New("song not found")

// Repository interface defines methods for interacting with songs in database
type Repository interface {
	Update(id int, song models.Song) (models.Song, error)
	Delete(id int) error
}

// Repo struct implements Repository interface and interacts with postgresql database using connection pool
type Repo struct {
	db database.Database
}

// New creates new Repo instance, taking database connection pool as parameter
func New(db database.Database) *Repo {
	return &Repo{db: db}
}

// Update modifies existing song in database by ID, and returns updated song
// If song with given ID not found, returns ErrSongNotFound
func (r *Repo) Update(id int, song models.Song) (models.Song, error) {
	query := `UPDATE songs SET group = $1, song = $2, release_date = $3, text = $4, link = $5, updated_at = NOW() 
             WHERE id = $6 RETURNING id, group, song, release_date, text, link, created_at, updated_at`
	ctx := context.Background()

	// Execute query and scan result into song object
	err := r.db.GetPool().QueryRow(ctx, query, song.Group, song.Title, song.ReleaseDate, song.Text, song.Link, id).
		Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt)
	if err != nil {
		// If no rows returned, return ErrSongNotFound
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Song{}, ErrSongNotFound
		}
		return models.Song{}, err
	}
	return song, nil
}

// Delete removes song from database by ID
// If song with given ID not found, returns ErrSongNotFound
func (r *Repo) Delete(id int) error {
	query := `DELETE FROM songs WHERE id = $1`
	ctx := context.Background()

	// Execute delete query and check how many rows were affected
	result, err := r.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		return err
	}

	// If no rows affected, return ErrSongNotFound
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSongNotFound
	}
	return nil
}
