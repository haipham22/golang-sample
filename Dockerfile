FROM golang:1.24-bullseye AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init --output ./internal/api/swagger --generalInfo ./internal/api/routes.go && \
    go build -v -o golang-sample

FROM debian:bullseye-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/golang-sample /app/golang-sample

CMD ["/app/golang-sample"]
