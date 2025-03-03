.PHONY: all run test docker lint migration seed gci format

# По умолчанию запускается команда run
all: docker migration seed start

run: docker start

start:
	@echo "Run cactus..."
	go run ./cmd/cactus/main.go &

	@echo "Waiting for cactus API..."
	while ! curl -s http://localhost:8080/healthz > /dev/null; do sleep 1; done
	@echo "Cactus API is ready"

	@echo "Run smtp-worker..."
	go run ./cmd/smtp-worker/ &
	@echo "Run telegram-worker..."
	go run ./cmd/telegram-worker/ &

docker:
	@echo "Run docker-compose..."
	docker-compose up -d --build

seed:
	@echo "Run seeding..."
	go run ./cmd/seeding/main.go

migration:
	@echo "Run migration..."
	dbmate --env-file ".env.dbmate.local" up

test:
	@echo "Run test..."
	CGO_ENABLED=1 go test -race -count 100 ./internal/...

lint:
	@echo "Run golangci-lint..."
	golangci-lint run ./...

gci:
	@echo "Run format imports..."
	gci write --skip-generated -s standard -s default -s localmodule .

format:
	@echo "Run format files..."
	gofumpt -w .
