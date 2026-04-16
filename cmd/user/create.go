package user

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (u *User) Create(cmd *cobra.Command) error {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	displayName, _ := cmd.Flags().GetString("name")

	if email == "" {
		return fmt.Errorf("email is required (--email)")
	}
	if password == "" {
		return fmt.Errorf("password is required (--password)")
	}

	if displayName == "" {
		for i, ch := range email {
			if ch == '@' {
				displayName = email[:i]
				break
			}
		}
	}

	usr, err := u.svc.Create(cmd.Context(), email, password, displayName, "")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Println("[OK] User created")
	fmt.Printf("  ID: %d\n", usr.ID)
	fmt.Printf("  Email: %s\n", usr.Email)
	fmt.Printf("  Display Name: %s\n", usr.DisplayName)
	fmt.Printf("  Status: %s\n", usr.Status)

	return nil
}
