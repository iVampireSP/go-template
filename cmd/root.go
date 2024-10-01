package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use: "rag",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
			return
		}
	},
}
