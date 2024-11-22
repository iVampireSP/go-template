package cmd

import (
	"context"
	"fmt"
	"go-template/internal/migrations"
	"os"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(migrateCommand, createGoMigrateCommand)
}

var migrateCommand = &cobra.Command{
	Use:   "goose [command]",
	Short: "goose 迁移，用法 <command>",
	Long:  "适用于生产环境的数据库迁移",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		RunMigrate(args)
	},
}

var createGoMigrateCommand = &cobra.Command{
	Use:   "create-migrate",
	Short: "新建 go 迁移",
	Long:  "新建 goose 的 go 迁移。",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		name := args[0]

		createGooseMigration(name)
	},
}

// RunMigrate 为数据库函数
func RunMigrate(args []string) {
	app, err := CreateApp()
	if err != nil {
		panic(err)
	}

	migrations.Config = app.Config

	goose.SetBaseFS(migrations.MigrationFS)

	err = goose.SetDialect("postgres")
	if err != nil {
		panic(err)
	}

	db, err := app.GORM.DB()
	if err != nil {
		panic(err)
	}

	command := args[0]

	var arguments []string
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	err = goose.RunContext(context.Background(), command, db, ".", arguments...)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
}

func createGooseMigration(name string) {
	// 在 internal/migrations 目录下新建一个迁移文件
	// 文件名为 yyyy-mm-dd-hh-mm-ss-<name>.go
	month := int(time.Now().Month())
	monthString := fmt.Sprintf("%d", month)
	if month < 10 {
		// 转 string
		monthString = "0" + monthString
	}

	day := time.Now().Day()
	dayString := fmt.Sprintf("%d", day)
	if day < 10 {
		dayString = "0" + dayString
	}

	hour := time.Now().Hour()
	hourString := fmt.Sprintf("%d", hour)
	if hour < 10 {
		hourString = "0" + hourString
	}

	minute := time.Now().Minute()
	minuteString := fmt.Sprintf("%d", minute)
	if minute < 10 {
		minuteString = "0" + minuteString
	}

	// 秒
	second := time.Now().Second()
	secondString := fmt.Sprintf("%d", second)
	if second < 10 {
		secondString = "0" + secondString
	}

	funcName := fmt.Sprintf("%d%s%s%s%s%s", time.Now().Year(), monthString, dayString, hourString, minuteString, secondString)
	fileName := fmt.Sprintf("%s_%s.go", funcName, name)

	// 模板内容
	var template = `package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up<FuncName>, Down<FuncName>)
}

func Up<FuncName>(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "UPDATE users SET username='admin' WHERE username='root';")
	return err
}

func Down<FuncName>(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "UPDATE users SET username='root' WHERE username='admin';")
	return err
}
`

	template = strings.ReplaceAll(template, "<FuncName>", funcName+name)
	err := os.WriteFile("internal/migrations/"+fileName, []byte(template), 0644)
	if err != nil {
		panic(fmt.Sprintf("failed creating migration file: %v", err))
	}

}
