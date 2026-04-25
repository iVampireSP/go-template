package command

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// Fresh drops all tables and re-runs all migrations from scratch.
func (m *Migrate) Fresh(cmd *cobra.Command) error {
	force, _ := cmd.Flags().GetBool("force")

	if !confirmDangerousOperation("drop ALL tables and rebuild database", force) {
		fmt.Println("Operation cancelled.")
		return nil
	}

	db, err := m.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	dbName := m.cfg.Name

	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE'
	`, dbName)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}
	rows.Close()

	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("[OK] Database rebuilt")
	return printVersion(db)
}
