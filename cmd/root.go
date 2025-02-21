package cmd

import (
	"github.com/spf13/cobra"
	"go-template/internal/infra"
)

var app = newApp()

func newApp() *infra.Application {
	a, err := CreateApp()
	if err != nil {
		panic(err)
	}

	return a
}

var RootCmd = &cobra.Command{
	Use: "app",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
		}
	},
}
