package migrations

import (
	"embed"
	"go-template/internal/base/conf"
)

//go:embed *.sql
var MigrationFS embed.FS

var Config *conf.Config
