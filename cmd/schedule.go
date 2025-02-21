package cmd

import (
	"go-template/internal/infra"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(scheduleCmd)
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule commands",
	Long:  `Schedule commands`,
	Run: func(cmd *cobra.Command, args []string) {
		app, err := CreateApp()
		if err != nil {
			panic(err)
		}

		runSchedule(app)
	},
}

func runSchedule(app *infra.Application) {
	// var wg sync.WaitGroup

	// var ctx = context.Background()

	// wg.Add(1)
	// // 启动一个定时器
	// go func() {

	// }()

	// wg.Wait()
}
