.PHONY: install-deps generate swag gorm proto buf

install-deps:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	echo "You need install altas manually, https://atlasgo.io/guides/evaluation/install"
	echo "You need install buf manually, https://github.com/bufbuild/buf"

gen:
	go generate ./...

buf:
	buf dep update

proto:
	buf generate

rehash:
	atlas migrate hash --dir file://migrations