package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(permissionManagement)
}

// permissionManagement 可以管理用户的权限，添加、删除、修改等
var permissionManagement = &cobra.Command{
	Use:   "permissions",
	Short: "Manage user permissions",
	Long:  "This command allows you to manage user permissions, including adding, and deleting permissions.",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
		}

		app, err := CreateApp()
		if err != nil {
			panic(err)
			return
		}

		add, _ := cmd.Flags().GetString("add")

		remove, _ := cmd.Flags().GetString("remove")

		user, _ := cmd.Flags().GetString("user")
		scope, _ := cmd.Flags().GetString("scope")
		if scope == "" {
			panic("scope is required")
		}

		if add != "" {
			fmt.Printf("Adding permission '%s'\n", add)

			success, err := app.Service.Auth.Enforcer.AddPermissionForUser(user, scope, add)
			//success, err := app.Service.Auth.Enforcer.AddPolicy(user, scope, add)

			if err != nil {
				panic(err)
				return
			}

			if success {
				fmt.Println("Permission added successfully.")
			} else {
				fmt.Println("Failed to add permission.")
			}

		} else if remove != "" {
			fmt.Printf("Removing permission '%s'\n", remove)

			success, err := app.Service.Auth.Enforcer.DeleteRoleForUser(user, remove)
			if err != nil {
				panic(err)
				return
			}

			if success {
				fmt.Println("Permission removed successfully.")
			} else {
				fmt.Println("Failed to remove permission.")
			}

		}

		err = app.Service.Auth.Enforcer.SavePolicy()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	permissionManagement.Flags().StringP("add", "a", "", "Add a permission to a user")
	permissionManagement.Flags().StringP("remove", "r", "", "Remove a permission from a user")
	permissionManagement.Flags().StringP("user", "u", "", "Specify the user ID")
	permissionManagement.Flags().StringP("scope", "s", "", "Specify the permission scope.")
}
