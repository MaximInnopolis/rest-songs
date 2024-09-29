package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"rest-songs/internal/app/models"
	"rest-songs/internal/app/repository/postgresql"
)

var ErrPageOutOfBounds = errors.New("page out of bounds")

type Service interface {
	GetSongsWithFilter(filter models.SongFilters, page, pageSize int) ([]models.Song, error)
	GetSongText(id, page, pageSize int) ([]string, error)
	UpdateSongById(id int, song models.Song) (models.Song, error)
	DeleteSongById(id int) error
	CreateSong(group, song string) (models.Song, error)
	FetchSongDetails(group, song string) (*models.SongDetail, error)
}

type SongService struct {
	repo        postgresql.Repository
	externalAPI string
	logger      *logrus.Logger
}

func New(repo postgresql.Repository, logger *logrus.Logger) *SongService {
	return &SongService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SongService) GetSongsWithFilter(filter models.SongFilters, page, pageSize int) ([]models.Song, error) {
	return s.repo.GetWithFilter(filter, page, pageSize)
}

func (s *SongService) GetSongText(id, page, pageSize int) ([]string, error) {
	s.logger.Infof("GetSongText[service]: Получение текста песни ID: %d, страница: %d, размер страницы: %d", id, page, pageSize)
	song, err := s.repo.GetById(id)
	if err != nil {
		s.logger.Errorf("GetSongText[service]: Ошибка получения песни по ID %d: %v", id, err)
		return nil, err
	}

	verses := strings.Split(song.Text, "\n\n")

	// Calculate pagination boundaries
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(verses) {
		s.logger.Warnf("GetSongText[service]: Страница %d выходит за пределы текста", page)
		return nil, ErrPageOutOfBounds
	}

	if end > len(verses) {
		end = len(verses)
	}

	// Return appropriate verses for requested page
	s.logger.Infof("GetSongText[service]: Успешно получено %d строк текста", end-start)
	return verses[start:end], nil
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

func (s *SongService) CreateSong(group, song string) (models.Song, error) {
	s.logger.Infof("CreateSong[service]: Создание песни группы: %s, название: %s", group, song)
	songDetails, err := s.FetchSongDetails(group, song)
	if err != nil {
		s.logger.Errorf("CreateSong[service]: Ошибка получения деталей песни через API: %v", err)
		return models.Song{}, err
	}

	// Parse release date from string to time.Time format
	releaseDate, err := time.Parse("02.01.2006", songDetails.ReleaseDate)
	if err != nil {
		s.logger.Errorf("CreateSong[service]: Ошибка парсинга даты релиза: %v", err)
		return models.Song{}, err
	}

	newSong := models.Song{
		Group:       group,
		Title:       song,
		ReleaseDate: releaseDate,
		Text:        songDetails.Text,
		Link:        songDetails.Link,
	}

	createdSong, err := s.repo.Create(newSong)
	if err != nil {
		s.logger.Errorf("CreateSong[service]: Ошибка создания песни в базе: %v", err)
		return models.Song{}, err
	}

	s.logger.Infof("CreateSong[service]: Песня успешно создана: %+v", createdSong)
	return createdSong, nil
}

func (s *SongService) FetchSongDetails(group, song string) (*models.SongDetail, error) {
	s.logger.Infof("FetchSongDetails[service]: Получение деталей песни через API для группы: %s, песни: %s", group, song)
	url := s.externalAPI + "/info?group=" + group + "&song=" + song

	resp, err := http.Get(url)
	if err != nil {
		s.logger.Errorf("FetchSongDetails[service]: Ошибка отправки запроса к API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("FetchSongDetails[service]: API вернул ошибку: %s", resp.Status)
		return nil, errors.New("API вернул ошибку: " + resp.Status)
	}

	// Парсинг ответа API
	var songDetails models.SongDetail
	if err = json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		s.logger.Errorf("FetchSongDetails[service]: Ошибка парсинга ответа от API: %v", err)
		return nil, err
	}

	s.logger.Infof("FetchSongDetails[service]: Успешно получены детали песни через API")
	return &songDetails, nil
}
