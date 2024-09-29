#!/bin/sh

#until pg_isready -h postgres -p 5432 -U postgres -d restSongs; do
#  echo "Waiting for database..."
#  sleep 2
#done

#goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restSongs?sslmode=disable" status

#goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restSongs?sslmode=disable" up


GOOSE_DRIVER=postgres GOOSE_DBSTRING="host=localhost port=5432 user=postgres password=password dbname=restSongs sslmode=disable" goose -dir ./migrations up
