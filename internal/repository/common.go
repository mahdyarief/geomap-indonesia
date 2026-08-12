package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// poolQuerier is the subset of *pgxpool.Pool used by repositories that do not
// need the full pool API. It allows easy test substitution.
type poolQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
