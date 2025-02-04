.PHONY: all run test lint

# По умолчанию запускается команда run
all: run

run:
	@echo "Run docker-compose..."
	docker-compose up -d --build
	@echo "Run migration..."
	dbmate --env-file ".env.dbmate.local" up
	@echo "Run seeding..."
	go run ./cmd/seeding/main.go
	@echo "Run cactus..."
	go run ./cmd/cactus/main.go
	@echo "Run email-worker..."
	go run ./cmd/emai-worker/

test:
	@echo "Run test..."
	CGO_ENABLED=1 go test -race -count 100 ./internal/...

lint:
	@echo "Run golangci-lint..."
	golangci-lint run ./...
