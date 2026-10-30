.PHONY: up down test lint

up:
	docker compose up --build

down:
	docker compose down

test:
	go test -race ./...

lint:
	go vet ./...