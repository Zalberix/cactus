FROM golang:1.22

WORKDIR /src

RUN go install github.com/cosmtrek/air@latest

COPY ..cmd/cactus /src/cmd/cactus
COPY ./go.mod /src/go.mod
COPY ./config/dev.yaml /src/config/dev.yaml

RUN go mod download

EXPOSE 8080 8080

CMD ["air", "-c", ".air.toml"]
