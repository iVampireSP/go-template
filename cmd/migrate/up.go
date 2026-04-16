package migrate

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// Up applies pending database migrations.
func (m *Migrate) Up(cmd *cobra.Command) error {
	if !hasMigrations() {
		fmt.Println("[OK] No migration files found, nothing to apply")
		return nil
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	args := cmd.Flags().Args()
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid step count: %s", args[0])
		}
		for range n {
			if err := goose.UpByOne(db, "."); err != nil {
				if errors.Is(err, goose.ErrNoNextVersion) {
					break
				}
				return fmt.Errorf("migration failed: %w", err)
			}
		}
		fmt.Printf("[OK] Applied up to %d migration(s)\n", n)
	} else {
		if err := goose.Up(db, ".", goose.WithAllowMissing()); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		fmt.Println("[OK] All migrations applied")
	}

	return printVersion(db)
}
