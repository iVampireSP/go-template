package orm

import (
	"database/sql"
	entsql "entgo.io/ent/dialect/sql"
	"fmt"
	"time"

	"github.com/google/wire"
	_ "github.com/lib/pq"
	"go-template/ent"
	"go-template/internal/infra/conf"
)

const (
	dbDriver = "postgres"
)

var (
	internalDB *sql.DB
)

func DSN(config *conf.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Database.Host, config.Database.User, config.Database.Password, config.Database.Name, config.Database.Port, config.Database.SSLMode, config.Database.TimeZone)
}

func DSNWithDriver(config *conf.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
		config.Database.SSLMode,
		config.Database.TimeZone)
}

var ProviderSet = wire.NewSet(
	NewSqlDB,
	NewEnt,
)

func NewSqlDB(config *conf.Config) *sql.DB {
	if internalDB != nil {
		return internalDB
	}

	db, err := sql.Open("postgres", DSNWithDriver(config))

	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Hour)

	internalDB = db

	return db
}

func NewEnt(config *conf.Config) *ent.Client {
	if internalDB == nil {
		NewSqlDB(config)
	}

	drv := entsql.OpenDB(dbDriver, internalDB)

	client := ent.NewClient(ent.Driver(drv))

	if config.Debug.Enabled {
		return client.Debug()
	}

	return client
}
