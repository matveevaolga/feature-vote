include .env
export

.PHONY: run build test docker-up docker-down docker-logs migrate-create migrate-up migrate-down seed clean

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./seeds

test:
	go test -v -race ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	migrate -path migrations -database ${CONN_STRING} up

migrate-down:
	migrate -path migrations -database ${CONN_STRING} down

seed:
	go run ./seeds

clean:
	rm -rf bin/