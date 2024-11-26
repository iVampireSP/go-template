.PHONY: install-deps generate swag gorm proto

install-deps:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	echo "You need install buf manually, https://github.com/bufbuild/buf"

generate:
	go generate ./...

swag:
	swag init -g main.go --parseDependency

gorm:
	cd hack/gorm-gen && go run .

proto:
	buf generate