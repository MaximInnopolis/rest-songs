# build docker image
build:
	docker-compose build

up-all:
	docker-compose up -d postgres app

down:
	docker-compose down

# run docker image
up-db:
	docker-compose up -d postgres

stop-db:
	docker-compose stop postgres

start-db:
	docker-compose start postgres

down-db:
	docker-compose down postgres


up-service:
	docker-compose up -d app --build

stop-service:
	docker-compose stop app

start-service:
	docker-compose start app

down-service:
	docker-compose down app


# Migration
migrate:
	./migration.sh up