# Build arguments for easier maintenance
ARG GO_VERSION=1.24
ARG DEBIAN_VERSION=bullseye
ARG APP_NAME=golang-sample
ARG WORK_DIR=/app

# Stage 1: Build the application
FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS builder
ARG WORK_DIR
WORKDIR ${WORK_DIR}

# Install tools first - rarely changes
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy only dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code that's needed for swagger generation
COPY . .

RUN swag init \
    --output ./internal/api/swagger \
    --generalInfo ./internal/api/routes.go || exit 0

RUN go build -v -o "${APP_NAME}"

# Stage 2: Create minimal runtime image
FROM debian:${DEBIAN_VERSION}-slim AS runtime
ARG WORK_DIR
ARG APP_NAME

# Install required certificates
RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Setup application
WORKDIR ${WORK_DIR}
COPY --from=builder ${WORK_DIR}/${APP_NAME} ${WORK_DIR}/${APP_NAME}

# Run the application
CMD ["${WORK_DIR}/${APP_NAME}"]
