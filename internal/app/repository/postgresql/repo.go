package postgresql

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/database"
)

var ErrSongNotFound = errors.New("song not found")

// Repository interface defines methods for interacting with songs in database
type Repository interface {
	GetWithFilter(filter models.SongFilters, page, pageSize int) ([]models.Song, error)
	GetById(id int) (models.Song, error)
	Update(id int, song models.Song) (models.Song, error)
	Delete(id int) error
	Create(song models.Song) (models.Song, error)
}

// Repo struct implements Repository interface and interacts with postgresql database using connection pool
type Repo struct {
	db     database.Database
	logger *logrus.Logger
}

// New creates new Repo instance, taking database connection pool and logger as parameters
func New(db database.Database, logger *logrus.Logger) *Repo {
	return &Repo{
		db:     db,
		logger: logger,
	}
}

func (r *Repo) GetWithFilter(filter models.SongFilters, page, pageSize int) ([]models.Song, error) {
	r.logger.Infof("GetWithFilter[repo]: Получение песен с фильтром: %+v, страница: %d, размер страницы: %d", filter, page, pageSize)

	query := `SELECT id, "group", song, release_date, text, link, created_at, updated_at 
           FROM songs WHERE 1=1` // Where 1=1 for filtering logic, so that further conditions also consider

	var songs []models.Song
	var args []interface{}
	argIndex := 1

	if filter.Group != "" {
		query += ` AND "group" = $` + strconv.Itoa(argIndex)
		args = append(args, filter.Group)
		argIndex++
	}

	if filter.Title != "" {
		query += ` AND song = $` + strconv.Itoa(argIndex)
		args = append(args, filter.Title)
		argIndex++
	}

	if !filter.ReleaseDate.IsZero() {
		query += ` AND release_date = $` + strconv.Itoa(argIndex)
		args = append(args, filter.ReleaseDate)
		argIndex++
	}

	// Add pagination
	offset := (page - 1) * pageSize
	query += ` ORDER BY release_date DESC LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
	args = append(args, pageSize, offset)

	ctx := context.Background()
	r.logger.Debugf("GetWithFilter[repo]: SQL запрос: %s, параметры: %+v", query, args)

	// Execute query and iterate over result rows
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.logger.Errorf("GetWithFilter[repo]: Ошибка выполнения SQL запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Scan each row into Song object and append to songs slice
	for rows.Next() {
		var song models.Song
		err = rows.Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate,
			&song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt)
		if err != nil {
			r.logger.Errorf("GetWithFilter[repo]: Ошибка сканирования строки: %v", err)
			return nil, err
		}
		songs = append(songs, song)
	}

	// Check for any error that occurred during iteration over rows
	if rows.Err() != nil {
		r.logger.Errorf("GetWithFilter[repo]: Ошибка при итерации по строкам: %v", rows.Err())
		return nil, rows.Err()
	}

	r.logger.Infof("GetWithFilter[repo]: Успешно получено %d песен", len(songs))
	return songs, nil
}

func (r *Repo) GetById(id int) (models.Song, error) {
	r.logger.Infof("GetById[repo]: Получение песни по ID: %d", id)

	query := `SELECT id, "group", song, release_date, text, link, created_at, updated_at FROM songs WHERE id = $1`
	var song models.Song
	ctx := context.Background()

	// Execute query and scan result into Song object
	err := r.db.GetPool().QueryRow(ctx, query, id).
		Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt)
	if err != nil {
		// If no rows returned, return ErrSongNotFound.
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("GetById[repo]: Песня с ID %d не найдена", id)
			return models.Song{}, ErrSongNotFound
		}
		r.logger.Errorf("GetById[repo]: Ошибка получения песни по ID %d: %v", id, err)
		return models.Song{}, err
	}

	r.logger.Infof("GetById[repo]: Успешно получена песня: %+v", song)
	return song, nil
}

// Update modifies existing song in database by ID, and returns updated song
// If song with given ID not found, returns ErrSongNotFound
func (r *Repo) Update(id int, song models.Song) (models.Song, error) {
	r.logger.Infof("Update[repo]: Обновление песни по ID: %d, данные: %+v", id, song)

	query := `UPDATE songs SET "group" = $1, song = $2, release_date = $3, text = $4, link = $5, updated_at = NOW() 
             WHERE id = $6 RETURNING id, "group", song, release_date, text, link, created_at, updated_at`
	ctx := context.Background()

	// Execute query and scan result into song object
	err := r.db.GetPool().QueryRow(ctx, query, song.Group, song.Title, song.ReleaseDate, song.Text, song.Link, id).
		Scan(&song.ID, &song.Group, &song.Title, &song.ReleaseDate, &song.Text, &song.Link, &song.CreatedAt, &song.UpdatedAt)
	if err != nil {
		// If no rows returned, return ErrSongNotFound
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("Update[repo]: Песня с ID %d не найдена для обновления", id)
			return models.Song{}, ErrSongNotFound
		}
		r.logger.Errorf("Update[repo]: Ошибка обновления песни по ID %d: %v", id, err)
		return models.Song{}, err
	}

	r.logger.Infof("Update[repo]: Успешно обновлена песня: %+v", song)
	return song, nil
}

// Delete removes song from database by ID
// If song with given ID not found, returns ErrSongNotFound
func (r *Repo) Delete(id int) error {
	r.logger.Infof("Delete[repo]: Удаление песни по ID: %d", id)

	query := `DELETE FROM songs WHERE id = $1`
	ctx := context.Background()

	// Execute delete query and check how many rows were affected
	result, err := r.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		r.logger.Errorf("Delete[repo]: Ошибка удаления песни по ID %d: %v", id, err)
		return err
	}

	// If no rows affected, return ErrSongNotFound
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warnf("Delete[repo]: Песня с ID %d не найдена для удаления", id)
		return ErrSongNotFound
	}
	r.logger.Infof("Delete[repo]: Успешно удалена песня по ID: %d", id)
	return nil
}

func (r *Repo) Create(song models.Song) (models.Song, error) {
	r.logger.Infof("Create[repo]: Создание новой песни: %+v", song)

	query := `INSERT INTO songs ("group", song, release_date, text, link, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`
	ctx := context.Background()

	// Execute query and scan returned ID, created_at, and updated_at into song object
	err := r.db.GetPool().QueryRow(ctx, query, song.Group, song.Title, song.ReleaseDate, song.Text, song.Link).
		Scan(&song.ID, &song.CreatedAt, &song.UpdatedAt)
	if err != nil {
		r.logger.Errorf("Create[repo]: Ошибка создания песни: %+v, ошибка: %v", song, err)
		return models.Song{}, err
	}
	r.logger.Infof("Create[repo]: Успешно создана песня: %+v", song)
	return song, nil
}
