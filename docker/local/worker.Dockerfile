FROM golang:1.23

WORKDIR /app

# Pre-download dependencies (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# CMD is overridden per-worker in docker-compose-workers.yml
