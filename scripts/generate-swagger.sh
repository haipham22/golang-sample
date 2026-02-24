#!/bin/bash

if ! command -v swag &> /dev/null
then
    echo "swag CLI could not be found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

swag init \
    --output ./internal/handler/rest/swagger \
    --generalInfo ./internal/handler/rest/routes.go \
    -td "[[,]]"

#    --parseDependency \
#    --parseInternal \
