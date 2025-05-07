# Golang Sample

## Directory structure

```
golang-sample
├── cmd
│   ├── api.go
│   ├── root.go
│   ├── sample.go
│   └── ...
├── internal // for apps
│   ├── api
│   │   ├── errors
│   │   ├── schemas
│   │   ├── storages
│   │   ├── transport
│   │   └── ...
│   └── another_packages
│       └── ....
├── pkg
│   ├── config
│   ├── databases
│   ├── errors
│   ├── healthcheck
│   ├── models
│   └── utils
│       ├── string
│       │   ├── hash.go
│       │   └── ...
│       └── ...
├── scripts
│   ├── golangci-lint.sh
│   └── ....
├── vendor
├── main.go
├── README.MD
└── ...
```

## Getting started

### Install Dependencies

From the project root, run:

```shell
go build ./...
go test ./...
go mod tidy
```


### Run dev

```shell
go run main.go api
```

### Generate swagger OpenAPI

Download Swag for Go
```shell
go install github.com/swaggo/swag/cmd/swag@latest
```

Format swag comments on Go code
```shell
go fmt
```

Generate swagger files (for local test)
```shell
./scripts/generate-swagger.sh
```

### TODO
- [ ] Add more tests
- [ ] Add ci lint && reviewdog
- [ ] Add more examples
