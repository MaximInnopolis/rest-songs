package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"rest-songs/internal/app/api"
	"rest-songs/internal/app/config"
	httpHandler "rest-songs/internal/app/http"
	"rest-songs/internal/app/repository/database"
	"rest-songs/internal/app/repository/postgresql"
)

// @title Songs API
// @version 1.0
// @description API for managing a song library
// @host localhost:8080
// @basePath /
// @schemes http

func main() {

	// Initialize logger
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	// Create config
	cfg, err := config.New()
	if err != nil {
		log.Errorf("Ошибка при чтении конфига: %v", err)
		os.Exit(1)
	}

	// Create a new connection pool to database
	pool, err := database.NewPool(cfg.DbUrl)
	if err != nil {
		log.Errorf("Ошибка при создании соединения к базе данных: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create a new Database with connection pool
	db := database.NewDatabase(pool)

	// Create a new repo with Database and logger
	repo := postgresql.New(*db, log)

	// Create a new service
	songService := api.New(repo, log)

	// Create Http handler
	handler := httpHandler.New(songService, cfg.ExternalAPI, log)

	// Init Router
	r := mux.NewRouter()

	handler.RegisterRoutes(r)

	// Start HTTP server
	if err = http.ListenAndServe(cfg.HttpPort, r); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	log.Infof("Сервер работает на порту: %s :", cfg.HttpPort)
}
