package main

import (
	"go-template/internal/entity"

	"gorm.io/gen"
)

// Dynamic SQL
//type Querier interface {
//	// SELECT * FROM @@table WHERE name = @name{{if role !=""}} AND role = @role{{end}}
//	FilterWithNameAndRole(name, role string) ([]gen.T, error)
//}

func main() {
	//app, err := cmd.CreateApp()
	//if err != nil {
	//	panic(err)
	//}
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/dao",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	//g.UseDB(app.GORM)

	g.ApplyBasic(
		entity.User{},
	)

	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	//g.ApplyInterface(func(Querier) {}, model.User{}, model.Company{})

	g.Execute()
}
