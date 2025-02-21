package cmd

import (
	"context"
	"fmt"
	"go-template/migrations"
	"log"
	"os"
	"strings"
	"time"

	"go-template/ent/migrate"
	"go-template/internal/infra/orm"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

const migrationPath = "migrations"

func init() {
	if app.Config.Debug.Enabled {
		RootCmd.AddCommand(createGoMigrateCommand, createEntMigrateCommand)
	}

	RootCmd.AddCommand(migrateCommand)
}

var migrateCommand = &cobra.Command{
	Use:   "goose [command]",
	Short: "goose <command>",
	Long:  "Run goose",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		RunMigrate(args)
	},
}

var createGoMigrateCommand = &cobra.Command{
	Use:   "create-go-migrate",
	Short: "create go migration",
	Long:  "create go migration using goose.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		name := args[0]
		createGooseMigration(name)
	},
}

var createEntMigrateCommand = &cobra.Command{
	Use:   "create-ent-migrate",
	Short: "create ent migration",
	Long:  "create ent migration using atlas.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		name := args[0]
		createEntMigration(name)
	},
}

// RunMigrate 为数据库函数
func RunMigrate(args []string) {
	migrations.Config = app.Config

	goose.SetBaseFS(migrations.MigrationFS)

	err := goose.SetDialect("postgres")
	if err != nil {
		panic(err)
	}

	command := args[0]

	var arguments []string
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	err = goose.RunContext(context.Background(), command, app.DB, ".", arguments...)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := app.DB.Close(); err != nil {
			panic(err)
		}
	}()
}

func createGooseMigration(name string) {
	// 在 migrationPath 目录下新建一个迁移文件
	// 文件名为 yyyy-mm-dd-hh-mm-ss-<name>.go
	month := int(time.Now().Month())
	monthString := fmt.Sprintf("%d", month)
	if month < 10 {
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

	second := time.Now().Second()
	secondString := fmt.Sprintf("%d", second)
	if second < 10 {
		secondString = "0" + secondString
	}

	funcName := fmt.Sprintf("%d%s%s%s%s%s", time.Now().Year(), monthString, dayString, hourString, minuteString, secondString)
	fileName := fmt.Sprintf("%s_%s.go", funcName, name)

	// 模板内容
	var template = `package ` + migrationPath + `

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
	err := os.WriteFile(migrationPath+"/"+fileName, []byte(template), 0644)
	if err != nil {
		panic(fmt.Sprintf("failed creating migration file: %v", err))
	}
}

func createEntMigration(name string) {
	app, err := CreateApp()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 创建一个本地迁移目录，用于理解 goose 迁移文件格式
	dir, err := sqltool.NewGooseDir(migrationPath)
	if err != nil {
		log.Fatalf("创建 atlas 迁移目录失败: %v", err)
	}

	// 迁移差异选项
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                          // 提供迁移目录
		schema.WithMigrationMode(schema.ModeInspect), // 提供迁移模式
		schema.WithDialect(dialect.Postgres),         // 使用 PostgreSQL 方言
		// schema.WithDropColumn(true),                 // 允许删除列
		// schema.WithDropIndex(true),                  // 允许删除索引
		// schema.WithForeignKeys(true),                // 处理外键
		// schema.WithGlobalUniqueID(true),             // 使用全局唯一ID
	}

	// 构建数据库 DSN
	dsn := orm.DSNWithDriver(app.Config)

	// 使用 Atlas 为 PostgreSQL 生成迁移文件
	err = migrate.NamedDiff(ctx, dsn, name, opts...)
	if err != nil {
		log.Fatalf("生成迁移文件失败: %v", err)
	}
}
