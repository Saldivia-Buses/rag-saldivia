// Package pagination provides helpers for paginated API responses.
package pagination

import (
	"net/http"
	"strconv"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 50
	MaxPageSize     = 100
)

// Params holds pagination parameters parsed from a request.
type Params struct {
	Page     int
	PageSize int
}

// Offset returns the SQL OFFSET for the current page.
func (p Params) Offset() int { return (p.Page - 1) * p.PageSize }

// Limit returns the SQL LIMIT (same as PageSize).
func (p Params) Limit() int { return p.PageSize }

// Parse extracts pagination params from query string (?page=&page_size=).
// Applies defaults and caps page_size to MaxPageSize.
func Parse(r *http.Request) Params {
	p := Params{Page: DefaultPage, PageSize: DefaultPageSize}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Page = n
		}
	}

	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.PageSize = n
		}
	}

	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}

	return p
}

// SetHeaders sets X-Page, X-Page-Size, and optionally X-Total-Count on the response.
func SetHeaders(w http.ResponseWriter, p Params, totalCount int) {
	w.Header().Set("X-Page", strconv.Itoa(p.Page))
	w.Header().Set("X-Page-Size", strconv.Itoa(p.PageSize))
	if totalCount >= 0 {
		w.Header().Set("X-Total-Count", strconv.Itoa(totalCount))
	}
}
