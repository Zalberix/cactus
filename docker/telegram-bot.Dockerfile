FROM golang:1.22

WORKDIR /src

COPY ./cmd/telegram-bot /src/cmd/telegram-bot
COPY ./go.mod /src/go.mod

ENV CGO_ENABLED=1

RUN go mod tidy

CMD ["go", "run", "/src/cmd/telegram-bot/main.go"]