FROM golang:1.23

WORKDIR /app

# Install air for live reload (via official install script — more reliable in Docker)
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b /usr/local/bin \
    && apt-get purge -y curl && apt-get autoremove -y && rm -rf /var/lib/apt/lists/*

# Pre-download dependencies (cache layer)
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.docker.toml"]
