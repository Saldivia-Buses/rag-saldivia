package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/items", nil)
	p := Parse(r)

	if p.Page != DefaultPage {
		t.Errorf("Page = %d, want %d", p.Page, DefaultPage)
	}
	if p.PageSize != DefaultPageSize {
		t.Errorf("PageSize = %d, want %d", p.PageSize, DefaultPageSize)
	}
}

func TestParse_ValidParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage int
		wantSize int
	}{
		{name: "page 2 size 20", query: "?page=2&page_size=20", wantPage: 2, wantSize: 20},
		{name: "page 1 size 1", query: "?page=1&page_size=1", wantPage: 1, wantSize: 1},
		{name: "page only", query: "?page=5", wantPage: 5, wantSize: DefaultPageSize},
		{name: "size only", query: "?page_size=25", wantPage: DefaultPage, wantSize: 25},
		{name: "max page size", query: "?page_size=100", wantPage: DefaultPage, wantSize: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			p := Parse(r)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantSize)
			}
		})
	}
}

func TestParse_CapsPageSize(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantSize int
	}{
		{name: "over max", query: "?page_size=200", wantSize: MaxPageSize},
		{name: "way over max", query: "?page_size=999999", wantSize: MaxPageSize},
		{name: "exactly max", query: "?page_size=100", wantSize: MaxPageSize},
		{name: "one over max", query: "?page_size=101", wantSize: MaxPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			p := Parse(r)

			if p.PageSize != tt.wantSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantSize)
			}
		})
	}
}

func TestParse_CapsPage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/items?page=99999", nil)
	p := Parse(r)

	if p.Page != MaxPage {
		t.Errorf("Page = %d, want %d (MaxPage)", p.Page, MaxPage)
	}
}

func TestParse_InvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage int
		wantSize int
	}{
		{name: "negative page", query: "?page=-1", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "zero page", query: "?page=0", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "negative size", query: "?page_size=-5", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "zero size", query: "?page_size=0", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "non-numeric page", query: "?page=abc", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "non-numeric size", query: "?page_size=xyz", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "float page", query: "?page=1.5", wantPage: DefaultPage, wantSize: DefaultPageSize},
		{name: "empty values", query: "?page=&page_size=", wantPage: DefaultPage, wantSize: DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			p := Parse(r)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantSize)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		wantOffset int
	}{
		{name: "page 1 size 50", page: 1, pageSize: 50, wantOffset: 0},
		{name: "page 2 size 50", page: 2, pageSize: 50, wantOffset: 50},
		{name: "page 3 size 20", page: 3, pageSize: 20, wantOffset: 40},
		{name: "page 10 size 100", page: 10, pageSize: 100, wantOffset: 900},
		{name: "page 1 size 1", page: 1, pageSize: 1, wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Params{Page: tt.page, PageSize: tt.pageSize}
			if got := p.Offset(); got != tt.wantOffset {
				t.Errorf("Offset() = %d, want %d", got, tt.wantOffset)
			}
		})
	}
}

func TestLimit(t *testing.T) {
	p := Params{Page: 3, PageSize: 25}
	if got := p.Limit(); got != 25 {
		t.Errorf("Limit() = %d, want 25", got)
	}
}

func TestSetHeaders(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		totalCount int
		wantPage   string
		wantSize   string
		wantTotal  string // empty string means header should not be set
	}{
		{
			name: "normal", page: 2, pageSize: 50, totalCount: 150,
			wantPage: "2", wantSize: "50", wantTotal: "150",
		},
		{
			name: "zero total", page: 1, pageSize: 20, totalCount: 0,
			wantPage: "1", wantSize: "20", wantTotal: "0",
		},
		{
			name: "negative total skips header", page: 1, pageSize: 50, totalCount: -1,
			wantPage: "1", wantSize: "50", wantTotal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			p := Params{Page: tt.page, PageSize: tt.pageSize}
			SetHeaders(w, p, tt.totalCount)

			if got := w.Header().Get("X-Page"); got != tt.wantPage {
				t.Errorf("X-Page = %q, want %q", got, tt.wantPage)
			}
			if got := w.Header().Get("X-Page-Size"); got != tt.wantSize {
				t.Errorf("X-Page-Size = %q, want %q", got, tt.wantSize)
			}
			gotTotal := w.Header().Get("X-Total-Count")
			if tt.wantTotal == "" {
				if gotTotal != "" {
					t.Errorf("X-Total-Count should not be set, got %q", gotTotal)
				}
			} else {
				if gotTotal != tt.wantTotal {
					t.Errorf("X-Total-Count = %q, want %q", gotTotal, tt.wantTotal)
				}
			}
		})
	}
}
