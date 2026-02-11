FROM golang:1.23 AS builder

WORKDIR /app

COPY ./go.mod /app/go.mod
RUN go mod download

CMD ["go", "run", "./app/cactus"]
