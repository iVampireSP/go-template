package command

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

// openDB creates a new database connection for migrations.
func (m *Migrate) openDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		m.cfg.User,
		m.cfg.Password,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.Name,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	_, _ = db.Exec("SET SESSION tidb_skip_isolation_level_check = 1")
	return db, nil
}

func printVersion(db *sql.DB) error {
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	if version == 0 {
		fmt.Println("  Current version: none")
	} else {
		fmt.Printf("  Current version: %d\n", version)
	}
	return nil
}
