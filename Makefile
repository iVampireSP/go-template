.PHONY: install-deps generate

install-deps:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

generate:
	go generate ./...

swag:
	swag init -g main.go --parseDependency
