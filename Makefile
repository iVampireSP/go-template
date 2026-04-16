.PHONY: gen build clean fmt test rename user_api admin_api \
        migrate-up migrate-down migrate-status migrate-fresh

# Code generation (ent ORM + autodi DI wiring)
gen:
	go generate ./...

# Build
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS    := -w -s \
	-X github.com/iVampireSP/go-template/pkg/version.Version=$(VERSION) \
	-X github.com/iVampireSP/go-template/pkg/version.Commit=$(COMMIT) \
	-X github.com/iVampireSP/go-template/pkg/version.BuildTime=$(BUILD_TIME)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/app .

test:
	go test ./...

clean:
	rm -rf bin/ tmp/

fmt:
	go fmt ./...

# Rename project module (interactive)
rename:
	go run ./hack/rename

# Run services
user_api:
	go run . serve api

admin_api:
	go run . serve admin

# ==================== Database Migrations ====================

migrate-up:
	go run . migrate up

migrate-down:
	go run . migrate down 1

migrate-status:
	go run . migrate status

migrate-fresh:
	go run . migrate fresh
