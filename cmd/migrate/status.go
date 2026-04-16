package migrate

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// Status displays the current state of all migrations.
func (m *Migrate) Status(cmd *cobra.Command) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Get all migrations from embedded FS
	allMigrations, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil {
		return fmt.Errorf("failed to collect migrations: %w", err)
	}

	if len(allMigrations) == 0 {
		fmt.Println("No migrations found.")
		return nil
	}

	// Get current version from database
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	// Print migrations table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "  VERSION\tNAME\tSTATUS")
	_, _ = fmt.Fprintln(w, "  -------\t----\t------")

	for _, mi := range allMigrations {
		name := strings.TrimSuffix(mi.Source, ".sql")
		parts := strings.SplitN(name, "_", 2)
		displayName := name
		if len(parts) == 2 {
			displayName = parts[1]
		}

		status := ""
		if mi.Version < currentVersion {
			status = "applied"
		} else if mi.Version == currentVersion {
			status = "current <--"
		} else {
			status = "pending"
		}
		_, _ = fmt.Fprintf(w, "  %d\t%s\t%s\n", mi.Version, displayName, status)
	}
	_ = w.Flush()

	return nil
}
