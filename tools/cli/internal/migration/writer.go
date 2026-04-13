package migration

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchWriter writes rows to PostgreSQL using INSERT ... ON CONFLICT DO NOTHING for idempotency.
type BatchWriter struct {
	pool     *pgxpool.Pool
	tenantID string
}

// NewBatchWriter creates a new batch writer.
func NewBatchWriter(pool *pgxpool.Pool, tenantID string) *BatchWriter {
	return &BatchWriter{pool: pool, tenantID: tenantID}
}

// WriteBatch inserts rows into the target table.
// columns lists the column names, conflictColumn is the unique column for ON CONFLICT.
// q is the executor (pool or transaction).
// Returns (rows written, error).
func (w *BatchWriter) WriteBatch(ctx context.Context, q querier, table string, columns []string, conflictColumn string, rows [][]any) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, strings.Join(columns, ",")))
	args := make([]any, 0, len(rows)*len(columns))
	for i, row := range rows {
		if len(row) != len(columns) {
			return 0, fmt.Errorf("row %d has %d values, expected %d columns", i, len(row), len(columns))
		}
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for j, val := range row {
			if j > 0 {
				sb.WriteString(",")
			}
			args = append(args, val)
			sb.WriteString(fmt.Sprintf("$%d", len(args)))
		}
		sb.WriteString(")")
	}
	if conflictColumn != "" {
		sb.WriteString(fmt.Sprintf(" ON CONFLICT (%s) DO NOTHING", conflictColumn))
	} else {
		sb.WriteString(" ON CONFLICT DO NOTHING")
	}

	tag, err := q.Exec(ctx, sb.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("write batch %s: %w", table, err)
	}
	return int(tag.RowsAffected()), nil
}

