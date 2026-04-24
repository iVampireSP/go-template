package main

import (
	"embed"

	migratecmd "github.com/iVampireSP/go-template/cmd/migrate"
	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/iVampireSP/go-template/internal/infra/i18n"
	"github.com/iVampireSP/go-template/internal/infra/tmpl"
)

//go:embed configs/*.yaml
var configsFS embed.FS

//go:embed all:database/migrations
var migrationsFS embed.FS

//go:embed lang
var langFS embed.FS

//go:embed all:templates
var templatesFS embed.FS

func init() {
	config.MustInitWithFS(configsFS, "configs")
	i18n.MustInitWithFS(langFS, "lang")
	migratecmd.MustInitWithFS(migrationsFS)
	tmpl.MustInitWithFS(templatesFS)
}
