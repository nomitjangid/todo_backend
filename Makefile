.PHONY: help build up down logs migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make build       - Build Docker images"
	@echo "  make up          - Start all services"
	@echo "  make down        - Stop all services"
	@echo "  make logs        - View logs"
	@echo "  make migrate-up  - Run database migrations"
	@echo "  make migrate-down- Rollback migrations"
	@echo "  make clean       - Remove all containers and volumes"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Services started successfully!"

down:
	docker-compose down

logs:
	docker-compose logs -f

migrate-up:
	docker-compose exec backend sh -c 'cd /root && migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" up'

migrate-down:
	docker-compose exec backend sh -c 'cd /root && migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" down'

clean:
	docker-compose down -v
	docker system prune -f