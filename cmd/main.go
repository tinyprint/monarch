package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"monarch"
	"os"

	"github.com/jackc/pgx/v5"
)

//go:embed help.txt
var helpText string

//go:embed help.create.txt
var helpCreateText string

//go:embed help.reapply.txt
var helpReapplyText string

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		fmt.Println(helpText)
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("provide a database URL via DATABASE_URL env var")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		return errors.New("provide a migrations path via MIGRATIONS_PATH env var")
	}

	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}

	migrator, err := monarch.NewMigrator(db, migrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	command := args[1]
	switch command {
	case "init":
		if len(args) > 2 {
			log.Fatalf("unexpected parameter %s", args[2])
		}
		return migrator.InitDirectory()
	case "migrate":
		if len(args) > 2 {
			log.Fatalf("unexpected parameter %s", args[2])
		}
		return migrator.Migrate(context.Background())
	case "create":
		if len(args) != 3 {
			fmt.Println(helpCreateText)
			os.Exit(1)
		}
		return migrator.Create(args[2])
	case "reapply":
		if len(args) != 3 {
			fmt.Println(helpReapplyText)
			os.Exit(1)
		}
		return migrator.Reapply(context.Background(), args[2])
	default:
		log.Fatalf("unknown command %s", command)
		return nil
	}
}
