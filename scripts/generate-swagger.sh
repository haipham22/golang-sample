#!/bin/bash

if ! command -v swag &> /dev/null
then
    echo "swag CLI could not be found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

swag init --output ./internal/api/swagger --generalInfo ./internal/api/routes.go
