.PHONY: build run test lint docker-up docker-down clean

build:
	go build -o kenko ./cmd/kenko

run: build
	./kenko -config configs/config.yaml

test:
	go test ./...

lint:
	golangci-lint run ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

clean:
	rm -f kenko
