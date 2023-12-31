package monarch

import (
	"context"
	"github.com/jackc/pgx/v5"
	"os"
	"strings"
	"testing"
)

func TestSupportedTypesReturnValuesAsExpected(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	if databaseURL == "" {
		t.Fatal("provide a database URL via DATABASE_URL env var")
	}

	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connection to database failed: %s", err)
	}

	err = runLua(ctx, db, runLuaConfig{
		file: "./test/lua_supported_types_return_values_as_expected.lua",
	})
	if err != nil {
		t.Fatalf("Lua test file failed with errors: %s", err)
	}
}

func TestUnfinishedQueryIteratorCausesError(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	if databaseURL == "" {
		t.Fatal("provide a database URL via DATABASE_URL env var")
	}

	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connection to database failed: %s", err)
	}

	err = runLua(ctx, db, runLuaConfig{
		file: "./test/lua_unfinished_query_iterator_causes_error.lua",
	})
	if err == nil {
		t.Fatalf("'conn busy' error expected; no error occurred")
	}

	if !strings.Contains(err.Error(), "conn busy") {
		t.Fatalf("Lua test file failed with unexpected error: %s", err)
	}
}

func TestUnfinishedQueryIteratorCanBeManuallyClosed(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	if databaseURL == "" {
		t.Fatal("provide a database URL via DATABASE_URL env var")
	}

	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connection to database failed: %s", err)
	}

	err = runLua(ctx, db, runLuaConfig{
		file: "./test/lua_unfinished_query_iterator_can_be_manually_closed.lua",
	})
	if err != nil {
		t.Fatalf("Lua test file failed with errors: %s", err)
	}
}
