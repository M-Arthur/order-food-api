run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run

build:
	go build -o bin/server ./cmd/server