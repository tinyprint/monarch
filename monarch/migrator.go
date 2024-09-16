package monarch

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

var regexpMigValidateName = regexp.MustCompile("^[a-zA-Z0-9_]+$")

type (
	MigrationFunc func(ctx context.Context, db *pgx.Conn, rollback bool) (bool, error)
)

type Migrator struct {
	db    *pgx.Conn
	model *model
	files *files
}

func NewMigrator(db *pgx.Conn, dir string) (*Migrator, error) {
	return &Migrator{
		db:    db,
		model: newModel(db),
		files: &files{directory: dir},
	}, nil
}

func (m *Migrator) InitDirectory() error {
	return m.files.initDirectory()
}

func (m *Migrator) Migrate(ctx context.Context) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	fmt.Println("running migrations")

	_, err = m.db.Exec(
		ctx,
		`
			CREATE TABLE IF NOT EXISTS migrations (
				migration_id varchar PRIMARY KEY,
				migrated_at timestamptz DEFAULT NOW() NOT NULL,
				last_applied_at timestamptz DEFAULT NOW() NOT NULL
			);
		`,
	)
	if err != nil {
		return err
	}

	migrated, skipped := 0, 0
	files, err := m.files.getMigrationFiles()
	if err != nil {
		return err
	}
	for _, name := range files {
		isMigrated, err := m.model.IsMigrated(ctx, name)
		if err != nil {
			return err
		}

		if isMigrated {
			skipped++
			if migrated > 0 {
				fmt.Printf("skipping previously migrated %s\n", name)
			}
			continue
		}

		if skipped > 0 && migrated == 0 {
			fmt.Printf("skipped %d previously migrated migrations\n", skipped)
		}

		fmt.Printf("running %s... ", name)
		migrationPath := m.files.migrationPath(name)
		err = runLua(ctx, m.db, runLuaConfig{file: migrationPath})
		if err != nil {
			return migrationFailed(
				fmt.Errorf("migration %s failed: %s", name, err.Error()),
			)
		}

		err = m.model.MarkAsMigrated(ctx, name)
		if err != nil {
			return migrationFailed(err)
		}

		fmt.Println("done")

		migrated++
	}

	if skipped > 0 && migrated == 0 {
		fmt.Printf("skipped %d previously migrated migrations\n", skipped)
	}

	return tx.Commit(ctx)
}

func (m *Migrator) Create(name string) error {
	if err := m.files.validateDirectory(); err != nil {
		return err
	}

	if !regexpMigValidateName.MatchString(name) {
		return fmt.Errorf(
			"%s is not a valid migration name (%s)",
			name,
			regexpMigValidateName.String(),
		)
	}

	err := m.files.createNewMigrationFile(name)
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) Reapply(ctx context.Context, name string) error {
	if err := m.files.validateDirectory(); err != nil {
		return err
	}

	migrationName := path.Base(name)
	if !strings.HasSuffix(migrationName, ".lua") {
		migrationName = migrationName + ".lua"
	}

	if !regexpMigMatchFileName.MatchString(migrationName) {
		return fmt.Errorf(
			"migration file %s does not exist",
			migrationName,
		)
	}

	isMigrated, err := m.model.IsMigrated(ctx, migrationName)
	if err != nil {
		return fmt.Errorf(
			"an error occured looking up %s migration's last run time: %s",
			migrationName,
			err,
		)
	} else if !isMigrated {
		return fmt.Errorf(
			"cannot reapply %s because it has not yet been ran; use `monarch migrate` first",
			migrationName,
		)
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	migrationPath := m.files.migrationPath(migrationName)
	err = runLua(ctx, m.db, runLuaConfig{
		file: migrationPath,
	})
	if err != nil {
		return migrationFailed(
			fmt.Errorf("migration %s failed: %s", migrationName, err.Error()),
		)
	}

	err = m.model.MarkAsReapplied(ctx, name)
	if err != nil {
		return migrationFailed(err)
	}

	fmt.Println("done")

	return tx.Commit(ctx)
}

func migrationFailed(err error) error {
	fmt.Println("failed")
	return err
}
