package user

import (
	usersvc "github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
	"github.com/spf13/cobra"
)

// User holds user management dependencies.
type User struct {
	app *container.Application
	svc *usersvc.User
}

// NewUser declares user command dependencies.
func NewUser(app *container.Application) *User {
	return &User{app: app}
}

// Command constructs the user command tree.
func (u *User) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "user", Short: "User management",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return u.app.Invoke(func(svc *usersvc.User) {
				u.svc = svc
			})
		},
	}

	create := &cobra.Command{Use: "create", Short: "Create user account",
		RunE: func(c *cobra.Command, _ []string) error { return u.Create(c) }}
	create.Flags().StringP("email", "e", "", "User email (required)")
	create.Flags().StringP("password", "p", "", "User password (required)")
	create.Flags().StringP("name", "n", "", "Display name")

	cmd.AddCommand(create)
	return cmd
}
