package importer

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mahdyarief/geomap-indonesia/internal/config"
	"github.com/mahdyarief/geomap-indonesia/internal/database"
)

// ConnectPool loads configuration and opens a verified connection pool.
func ConnectPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return database.Connect(ctx, cfg)
}

// CopyTable bulk-inserts rows into table using COPY. Rows must match the
// column list order; nil values become SQL NULL.
func CopyTable(ctx context.Context, pool *pgxpool.Pool, table string, cols []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := pool.CopyFrom(ctx, pgx.Identifier{table}, cols, pgx.CopyFromRows(rows))
	return err
}

// ExecBatch sends a batch of statements to the pool (auto-commit).
func ExecBatch(ctx context.Context, pool *pgxpool.Pool, batch *pgx.Batch) error {
	br := pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
