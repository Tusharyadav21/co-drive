.PHONY: all build check run dev migrate migrate-up migrate-down test clean

all: build

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

check:
	go vet ./...
	go test -v ./...
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

migrate: migrate-up

migrate-up: build
	./bin/migrate -direction up

migrate-down: build
	./bin/migrate -direction down

dev: migrate-up
	./bin/server

run: build
	./bin/server

test:
	go test -v ./...

clean:
	rm -rf bin/ data/