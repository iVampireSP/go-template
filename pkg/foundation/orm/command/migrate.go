package command

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/iVampireSP/go-template/pkg/foundation/lock"
	"github.com/iVampireSP/go-template/pkg/foundation/orm"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// globalFS 保存嵌入的迁移文件系统
var globalFS embed.FS

// MustInitWithFS 初始化迁移文件系统
func MustInitWithFS(migrationsFS embed.FS) {
	subFS, err := fs.Sub(migrationsFS, "database/migrations")
	if err != nil {
		panic(fmt.Sprintf("failed to get migrations sub-fs: %v", err))
	}
	goose.SetTableName("migrations")
	goose.SetBaseFS(subFS)
	if err := goose.SetDialect("tidb"); err != nil {
		panic(fmt.Sprintf("failed to set goose dialect: %v", err))
	}
	globalFS = migrationsFS
}

// Migrate provides database migration commands.
type Migrate struct {
	app    *container.Application
	cfg    orm.Config
	locker *lock.Locker
}

// NewMigrate creates a new Migrate command group.
func NewMigrate(app *container.Application) *Migrate {
	return &Migrate{app: app}
}

// Command returns the migrate command tree.
func (m *Migrate) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrate", Short: "Database migrations",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return m.app.Invoke(func(cfg orm.Config, locker *lock.Locker) {
				m.cfg = cfg
				m.locker = locker
			})
		},
	}

	up := &cobra.Command{Use: "up [N]", Short: "Apply migrations",
		RunE: func(c *cobra.Command, _ []string) error { return m.Up(c) }}
	down := &cobra.Command{Use: "down [N]", Short: "Rollback migrations (default: 1 step)",
		RunE: func(c *cobra.Command, _ []string) error { return m.Down(c) }}
	down.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	status := &cobra.Command{Use: "status", Short: "Show migration status",
		RunE: func(c *cobra.Command, _ []string) error { return m.Status(c) }}

	fresh := &cobra.Command{Use: "fresh", Short: "Drop all tables and re-run migrations (DANGEROUS)",
		RunE: func(c *cobra.Command, _ []string) error { return m.Fresh(c) }}
	fresh.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	cmd.AddCommand(up, down, status, fresh)

	return cmd
}

// hasMigrations 检查嵌入的迁移目录中是否存在 .sql 文件
func hasMigrations() bool {
	entries, err := fs.ReadDir(globalFS, "database/migrations")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			return true
		}
	}
	return false
}
