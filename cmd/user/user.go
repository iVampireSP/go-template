package user

import (
	usersvc "github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/spf13/cobra"
)

// User holds user management dependencies.
type User struct {
	svc *usersvc.User
}

// NewUser declares user command dependencies.
func NewUser(svc *usersvc.User) *User {
	return &User{
		svc: svc,
	}
}

// Command constructs the user command tree.
func (u *User) Command() *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "User management"}

	create := &cobra.Command{Use: "create", Short: "Create user account"}
	create.Flags().StringP("email", "e", "", "User email (required)")
	create.Flags().StringP("password", "p", "", "User password (required)")
	create.Flags().StringP("name", "n", "", "Display name")

	cmd.AddCommand(create)
	return cmd
}
