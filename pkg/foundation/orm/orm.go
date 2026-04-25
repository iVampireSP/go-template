package orm

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iVampireSP/go-template/ent"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// NewORM creates a new ent client for MySQL/TiDB.
func NewORM(cfg Config) *ent.Client {
	db, err := OpenDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	// Create an ent driver from the existing sql.DB
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	return client
}

// GetDSN returns the MySQL DSN string from config.
func GetDSN(cfg Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}

// OpenDB opens a MySQL database connection with OTel SQL tracing.
func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := otelsql.Open("mysql", GetDSN(cfg),
		otelsql.WithAttributes(
			semconv.DBSystemNameMySQL,
			semconv.DBNamespace(cfg.Name),
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent port exhaustion
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

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
