package command

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// Down rolls back database migrations.
func (m *Migrate) Down(cmd *cobra.Command) error {
	force, _ := cmd.Flags().GetBool("force")

	args := cmd.Flags().Args()
	n := 1
	if len(args) > 0 {
		var err error
		n, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid step count: %s", args[0])
		}
	}

	if !confirmDangerousOperation(fmt.Sprintf("rollback %d migration(s)", n), force) {
		fmt.Println("Operation cancelled.")
		return nil
	}

	db, err := m.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	for i := 0; i < n; i++ {
		if err := goose.Down(db, ".", goose.WithAllowMissing()); err != nil {
			if errors.Is(err, goose.ErrNoNextVersion) {
				break
			}
			return fmt.Errorf("rollback failed: %w", err)
		}
	}

	fmt.Printf("[OK] Rolled back %d migration(s)\n", n)
	return printVersion(db)
}
