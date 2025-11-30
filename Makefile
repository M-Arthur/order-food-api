run:
	docker compose -f deploy/docker-compose.yml up --build

build:
	docker compose build

destroy:
	docker compose -f deploy/docker-compose.yml down --volumes --remove-orphans

test:
	go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run
