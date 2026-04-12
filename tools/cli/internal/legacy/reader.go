package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// LegacyRow is a generic row from MySQL with column name → value mapping.
type LegacyRow map[string]any

// Int64 returns an int64 value or 0.
func (r LegacyRow) Int64(col string) int64 {
	v, ok := r[col]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case []byte:
		n, _ := strconv.ParseInt(string(val), 10, 64)
		return n
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	default:
		return 0
	}
}

// Int returns an int value.
func (r LegacyRow) Int(col string) int {
	return int(r.Int64(col))
}

// String returns a string value or empty string.
func (r LegacyRow) String(col string) string {
	v, ok := r[col]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// NullString returns a *string (nil if NULL or empty).
func (r LegacyRow) NullString(col string) *string {
	v, ok := r[col]
	if !ok || v == nil {
		return nil
	}
	s := r.String(col)
	if s == "" {
		return nil
	}
	return &s
}

// Decimal returns the string representation of a decimal value (for shopspring/decimal parsing).
func (r LegacyRow) Decimal(col string) string {
	v, ok := r[col]
	if !ok || v == nil {
		return "0"
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Reader reads batches from a single legacy MySQL table.
type Reader interface {
	// LegacyTable returns the MySQL table name this reader handles.
	LegacyTable() string
	// SDATable returns the target PostgreSQL table name.
	SDATable() string
	// Domain returns the migration domain (e.g., "catalog", "entity", "accounting").
	Domain() string
	// ReadBatch reads up to limit rows with key > resumeKey.
	// resumeKey="" means start from the beginning.
	// Returns rows, the key of the last row (for resume), and error.
	ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error)
}

// GenericReader implements Reader for tables with a single BIGINT primary key.
type GenericReader struct {
	DB          *sql.DB
	Table       string // legacy table name
	Target      string // SDA table name
	DomainName  string
	PKColumn    string
	Columns     string // SELECT columns
	ExtraWhere  string // optional WHERE clause (without WHERE keyword)
}

func (r *GenericReader) LegacyTable() string { return r.Table }
func (r *GenericReader) SDATable() string     { return r.Target }
func (r *GenericReader) Domain() string       { return r.DomainName }

func (r *GenericReader) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error) {
	lastID := int64(0)
	if resumeKey != "" {
		lastID, _ = strconv.ParseInt(resumeKey, 10, 64)
	}

	where := fmt.Sprintf("%s > ?", r.PKColumn)
	if r.ExtraWhere != "" {
		where = where + " AND " + r.ExtraWhere
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT ?",
		r.Columns, r.Table, where, r.PKColumn)

	rows, err := r.DB.QueryContext(ctx, query, lastID, limit)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", r.Table, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("columns %s: %w", r.Table, err)
	}

	var result []LegacyRow
	var lastKey int64
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, "", fmt.Errorf("scan %s: %w", r.Table, err)
		}
		row := make(LegacyRow, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
		lastKey = row.Int64(r.PKColumn)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate %s: %w", r.Table, err)
	}

	return result, strconv.FormatInt(lastKey, 10), nil
}

// CompositeKeyReader handles tables with composite primary keys.
type CompositeKeyReader struct {
	DB         *sql.DB
	Table      string
	Target     string
	DomainName string
	PKColumns  []string // e.g., ["FACTURA_ID", "TIPO_IVA"]
	Columns    string
	ExtraWhere string
}

func (r *CompositeKeyReader) LegacyTable() string { return r.Table }
func (r *CompositeKeyReader) SDATable() string     { return r.Target }
func (r *CompositeKeyReader) Domain() string       { return r.DomainName }

// ParseCompositeKey parses "COL1=val1,COL2=val2" into a map.
func ParseCompositeKey(key string) map[string]string {
	result := make(map[string]string)
	if key == "" {
		return result
	}
	for _, part := range strings.Split(key, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// FormatCompositeKey formats a map into "COL1=val1,COL2=val2".
func FormatCompositeKey(cols []string, row LegacyRow) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = fmt.Sprintf("%s=%s", col, row.String(col))
	}
	return strings.Join(parts, ",")
}

func (r *CompositeKeyReader) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error) {
	parsed := ParseCompositeKey(resumeKey)

	var where string
	var args []any
	if len(parsed) > 0 {
		// Build (col1, col2) > (?, ?) for composite key ordering
		colList := strings.Join(r.PKColumns, ", ")
		placeholders := make([]string, len(r.PKColumns))
		for i, col := range r.PKColumns {
			placeholders[i] = "?"
			args = append(args, parsed[col])
		}
		where = fmt.Sprintf("(%s) > (%s)", colList, strings.Join(placeholders, ", "))
	} else {
		where = "1=1"
	}
	if r.ExtraWhere != "" {
		where = where + " AND " + r.ExtraWhere
	}

	orderBy := strings.Join(r.PKColumns, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT ?",
		r.Columns, r.Table, where, orderBy)
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", r.Table, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("columns %s: %w", r.Table, err)
	}

	var result []LegacyRow
	var lastRow LegacyRow
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, "", fmt.Errorf("scan %s: %w", r.Table, err)
		}
		row := make(LegacyRow, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
		lastRow = row
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate %s: %w", r.Table, err)
	}

	lastKey := ""
	if lastRow != nil {
		lastKey = FormatCompositeKey(r.PKColumns, lastRow)
	}
	return result, lastKey, nil
}
