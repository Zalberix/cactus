.PHONY: all dev dev-down dev-logs run test docker docker-workers docker-down lint migration migrate-down migrate-status migrate-create gci format watch

# Локальная разработка в Docker
dev: docker-infra docker-workers
	@echo ""
	@echo "Stack running:"
	@echo "  API      -> http://localhost:8080"
	@echo "  Web      -> http://localhost:3000"
	@echo "  MailHog  -> http://localhost:8025"
	@echo "  RustFS   -> http://localhost:9001  (s3admin / s3admin)"
	@echo "  PgSQL    -> localhost:5432"
	@echo "  Redis    -> localhost:6379"

dev-down:
	@echo "Stopping workers..."
	docker compose -f ./docker-compose-workers.yml down
	@echo "Stopping infra..."
	docker compose -f ./docker-compose.yml down

# Внутренние цели
docker-infra:
	@echo "Starting infrastructure (db, redis, mailhog, rustfs, api, web)..."
	docker compose -f ./docker-compose.yml up -d --build

docker-workers:
	@echo "Starting workers (smtp, telegram, telegram-bot)..."
	docker compose -f ./docker-compose-workers.yml up -d --build

docker-down:
	docker compose -f ./docker-compose.yml down
	docker compose -f ./docker-compose-workers.yml down

# Устаревший запуск (вне Docker)
run: docker-infra start

start:
	@echo "Run cactus..."
	go run ./apps/core/cmd/cactus/main.go &

	@echo "Waiting for cactus API..."
	while ! curl -s http://localhost:8080/healthz > /dev/null; do sleep 1; done
	@echo "Cactus API is ready"

	@echo "Run smtp-worker..."
	go run ./cmd/smtp-worker/ &
	@echo "Run telegram-worker..."
	go run ./cmd/telegram-worker/ &

migration:
	@echo "Applying migrations..."
	go run ./cli/main.go migration up

migration-down:
	@echo "Rolling back last migration..."
	go run ./cli/main.go migration down

migration-status:
	@echo "Migration status..."
	go run ./cli/main.go migration status

migration-create:
	@echo "Creating migration: $(name)"
	go run ./cli/main.go migration create $(name)

test:
	@echo "Run test..."
	CGO_ENABLED=1 go test -race -count 100 ./apps/core/...

lint:
	@echo "Run golangci-lint..."
	golangci-lint run ./...

gci:
	@echo "Run format imports..."
	gci write --skip-generated -s standard -s default -s localmodule .

watch:
	@powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
		Write-Output 'Watching...'; \
	} else { \
		Write-Output 'Installing air...'; \
		go install github.com/air-verse/air@latest; \
		air; \
		Write-Output 'Watching...'; \
	}"

format:
	@echo "Run format files..."
	gofumpt -w .
