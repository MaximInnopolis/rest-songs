package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database struct holds the connection pool to postgresql database
type Database struct {
	pool *pgxpool.Pool
}

// GetPool returns connection pool instance
func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}

// NewDatabase creates new Database instance with connection pool
func NewDatabase(pool *pgxpool.Pool) *Database {
	return &Database{pool: pool}
}

// NewPool initializes new connection pool to the postgresql database using database url
// It connects to database in background and returns pool
func NewPool(dbUrl string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), dbUrl)
}
