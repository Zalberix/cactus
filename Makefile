.PHONY: all test lint gci format

test:
	@echo "Run test..."
	CGO_ENABLED=1 go test -race -count 100 ./apps/core/...

lint:
	@echo "Run golangci-lint..."
	golangci-lint run ./...

gci:
	@echo "Run format imports..."
	gci write --skip-generated -s standard -s default -s localmodule .

format:
	@echo "Run format files..."
	gofumpt -w .
