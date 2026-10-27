package testutil

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewTestDB(t testing.TB) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("TEST_DB_DSN"))
	if dsn == "" {
		t.Fatal("TEST_DB_DSN must be set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	t.Cleanup(func() {
		db.Close()
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	var dbName string
	if err := db.QueryRowContext(
		ctx,
		"SELECT current_database()",
	).Scan(&dbName); err != nil {
		t.Fatalf("verify test database: %v", err)
	}

	if dbName != "palge_test" {
		t.Fatalf(
			"expected database %q, got %q",
			"palge_test",
			dbName,
		)
	}

	return db
}

// func NewTestDB(t testing.TB) *pgxpool.Pool {
// 	t.Helper()

// 	dsn := strings.TrimSpace(os.Getenv("TEST_DB_DSN"))
// 	if dsn == "" {
// 		t.Fatal("TEST_DB_DSN must be set")
// 	}

// 	cfg, err := pgxpool.ParseConfig(dsn)
// 	if err != nil {
// 		t.Fatalf("parse TEST_DB_DSN: %v", err)
// 	}

// 	cfg.MaxConns = 5

// 	ctx, cancel := context.WithTimeout(
// 		context.Background(),
// 		5*time.Second,
// 	)
// 	defer cancel()

// 	db, err := pgxpool.NewWithConfig(ctx, cfg)
// 	if err != nil {
// 		t.Fatalf("create test database pool: %v", err)
// 	}

// 	t.Cleanup(func() {
// 		db.Close()
// 	})

// 	if err := db.Ping(ctx); err != nil {
// 		t.Fatalf("ping test database: %v", err)
// 	}

// 	var dbName string
// 	if err := db.QueryRow(
// 		ctx,
// 		"SELECT current_database()",
// 	).Scan(&dbName); err != nil {
// 		t.Fatalf("verify test database: %v", err)
// 	}

// 	if dbName != "palge_test" {
// 		t.Fatalf(
// 			"expected database %q, got %q",
// 			"palge_test",
// 			dbName,
// 		)
// 	}

// 	return db
// }
