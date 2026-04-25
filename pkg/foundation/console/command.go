package console

import "github.com/spf13/cobra"

// ConsoleCommand defines a command that can be registered with the application.
// This mirrors Laravel's Illuminate\Console\Command.
type ConsoleCommand interface {
	Command() *cobra.Command
}
