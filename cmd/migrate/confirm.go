package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/iVampireSP/go-template/internal/infra/config"
)

// confirmDangerousOperation prompts for confirmation in non-development environments.
// Returns true if the operation should proceed.
func confirmDangerousOperation(operation string, force bool) bool {
	if config.IsDevelopment() {
		return true
	}
	if force {
		return true
	}

	env := config.Env()
	fmt.Printf("\n[WARNING] You are about to %s in '%s' environment.\n", operation, env)
	fmt.Print("Are you sure you want to continue? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
