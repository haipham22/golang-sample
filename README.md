# paperback-vbook-converter

## Directory structure

```
.
├── main.go
├── .env
├── cmd
│   ├── root.go
│   └── another_command.go
├── pkg
│   ├── common
│   └── another_package
├── ...
...
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
go run main.go sample
```
