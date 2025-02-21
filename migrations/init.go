package migrations

import (
	"embed"
	"go-template/internal/infra/conf"
)

//go:embed *.sql
var MigrationFS embed.FS

var Config *conf.Config
