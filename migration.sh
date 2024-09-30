goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restSongs?sslmode=disable" status

goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restSongs?sslmode=disable" up