package monarch

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type model struct {
	db             *pgx.Conn
	migratedByName map[string]bool
}

func newModel(db *pgx.Conn) *model {
	return &model{db: db}
}

func (model *model) IsMigrated(ctx context.Context, id string) (bool, error) {
	if model.migratedByName == nil {
		err := model.loadMigratedByName(ctx)
		if err != nil {
			return false, err
		}
	}

	ran, exists := model.migratedByName[id]
	if !exists {
		return false, nil
	}

	return ran, nil
}

func (model *model) MarkAsMigrated(ctx context.Context, id string) error {
	_, err := model.db.Exec(
		ctx,
		`
			INSERT INTO migrations (migration_id)
			VALUES ($1)
		`,
		id,
	)

	return err
}

func (model *model) MarkAsReapplied(ctx context.Context, id string) error {
	_, err := model.db.Exec(
		ctx,
		`
			UPDATE migrations
			SET last_applied_at = NOW()
			WHERE migration_id = $1
		`,
		id,
	)

	return err
}

func (model *model) loadMigratedByName(ctx context.Context) error {
	model.migratedByName = make(map[string]bool)

	rows, err := model.db.Query(
		ctx,
		`
			SELECT migration_id
			FROM migrations
		`,
	)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return err
		}

		model.migratedByName[id] = true
	}

	return nil
}
