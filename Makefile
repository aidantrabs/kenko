.PHONY: build run test test-cover lint docker-up docker-down clean

build:
	go build -o kenko ./cmd/kenko

run: build
	./kenko -config configs/config.yaml

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

clean:
	rm -f kenko coverage.out
