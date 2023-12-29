package monarch

import (
	"context"
	"github.com/jackc/pgx/v5"
	"os"
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
		file: "./test/support_types_return_values_as_expected.lua",
	})
	if err != nil {
		t.Fatalf("Lua test file failed with errors: %s", err)
	}

}
