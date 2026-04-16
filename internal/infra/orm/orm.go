package orm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/iVampireSP/go-template/ent"
	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// NewORM creates a new ent client for MySQL/TiDB.
func NewORM() *ent.Client {
	db, err := OpenDB()
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	// Create an ent driver from the existing sql.DB
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	return client
}

// GetDSN returns the MySQL DSN string from config.
func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		config.String("database.app.user", "root"),
		config.String("database.app.password"),
		config.String("database.app.host", "localhost"),
		config.Int("database.app.port", 4000),
		config.String("database.app.name", "cloud"),
	)
}

// OpenDB opens a MySQL database connection with OTel SQL tracing.
func OpenDB() (*sql.DB, error) {
	db, err := otelsql.Open("mysql", GetDSN(),
		otelsql.WithAttributes(
			semconv.DBSystemNameMySQL,
			semconv.DBNamespace(config.String("database.app.name", "cloud")),
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent port exhaustion
	db.SetMaxOpenConns(config.Int("database.app.max_open_conns", 25))
	db.SetMaxIdleConns(config.Int("database.app.max_idle_conns", 5))
	db.SetConnMaxLifetime(time.Duration(config.Int("database.app.conn_max_lifetime_seconds", 300)) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.Int("database.app.conn_max_idle_time_seconds", 60)) * time.Second)

	return db, nil
}

// CloseEntClient closes the ent client connection.
func CloseEntClient(client *ent.Client) error {
	return client.Close()
}

// WithTx executes a function within a database transaction.
func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
