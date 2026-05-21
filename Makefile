SHELL := /bin/bash

.PHONY: help run build test test-v compose-up compose-down compose-logs db-shell tidy

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Corre la API en local (requiere .env)
	@set -a; source .env; set +a; go run ./cmd/api/...

build: ## Compila el binario en bin/api
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/api ./cmd/api/...

test: ## Corre todos los tests con race detector
	go test -race -cover ./...

test-v: ## Corre los tests en modo verbose
	go test -race -v -cover ./...

tidy: ## Actualiza dependencias
	go mod tidy && go mod verify

compose-up: ## Levanta todos los servicios con Podman
	podman compose up -d

compose-down: ## Detiene todos los servicios
	podman compose down

compose-logs: ## Sigue los logs de la API
	podman compose logs -f api

db-shell: ## Abre psql en el contenedor de la base de datos
	podman exec -it task-manager-db psql -U postgres -d taskmanager
