package monarch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func getManagementConnection(ctx context.Context) (*pgx.Conn, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("provide a database URL via DATABASE_URL env var")
	}

	db, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connection to database failed: %s", err)
	}

	return db, nil
}

func getTestConnection(ctx context.Context) (*pgx.Conn, *pgx.Conn, func(context.Context), error) {
	management, err := getManagementConnection(ctx)
	if err != nil {
		return nil, nil, func(context.Context) {}, err
	}

	testDBName := fmt.Sprintf("monarch_test_%s", time.Now().Format("20060102_150405_999999999"))
	_, err = management.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{testDBName}.Sanitize()))
	if err != nil {
		return nil, nil, func(context.Context) {}, fmt.Errorf("error creating test database `%s`: %w", testDBName, err)
	}

	testConfig := management.Config()
	testConfig.Database = testDBName

	db, err := pgx.ConnectConfig(ctx, testConfig)
	if err != nil {
		management.Close(ctx)
		return nil, nil, func(context.Context) {}, fmt.Errorf("error connecting to test database: %w", err)
	}

	assertDB, err := pgx.ConnectConfig(ctx, testConfig)
	if err != nil {
		management.Close(ctx)
		db.Close(ctx)
		return nil, nil, func(context.Context) {}, fmt.Errorf("error connecting to test database for assertions: %w", err)
	}

	return db, assertDB, func(cleanupCtx context.Context) {
		assertDB.Close(cleanupCtx)
		db.Close(cleanupCtx)
		_, err = management.Exec(cleanupCtx, fmt.Sprintf("DROP DATABASE %s", pgx.Identifier{testDBName}.Sanitize()))
		if err != nil {
			fmt.Printf("error dropping test database `%s`: %s", testDBName, err)
		}
		management.Close(cleanupCtx)
	}, nil
}

func TestMigrationsCanRunSuccessfully(t *testing.T) {
	ctx := context.Background()
	db, assertDB, cleanup, err := getTestConnection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(ctx)

	migrator, err := NewMigrator(db, "./test/working_migrations")
	if err != nil {
		t.Fatalf("error setting up migrator: %s", err)
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		t.Fatalf("error running migrations: %s", err)
	}

	id := 0
	row := assertDB.QueryRow(ctx, "SELECT id FROM test_table")
	err = row.Scan(&id)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected 'no rows error'; got: %s", err)
	}
}
