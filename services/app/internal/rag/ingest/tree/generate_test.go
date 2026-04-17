package tree

import (
	"encoding/json"
	"testing"
)

func TestBuildUnifiedText(t *testing.T) {
	pages := []ExtractionPage{
		{PageNumber: 1, Text: "# Title\nSome text"},
		{PageNumber: 2, Text: "More content"},
	}
	result := buildUnifiedText(pages)
	if len(result) == 0 {
		t.Fatal("empty unified text")
	}
	if !contains(result, "<page_1>") || !contains(result, "</page_2>") {
		t.Fatal("missing page markers")
	}
}

func TestBuildUnifiedText_WithImages(t *testing.T) {
	pages := []ExtractionPage{
		{
			PageNumber: 1,
			Text:       "Chapter 1",
			Images: []struct {
				Description   string   `json:"description"`
				Type          string   `json:"type"`
				ExtractedData []string `json:"extracted_data"`
			}{
				{Description: "Engine diagram", Type: "technical_diagram"},
			},
		},
	}
	result := buildUnifiedText(pages)
	if !contains(result, "[IMAGEN: Engine diagram]") {
		t.Fatal("missing image description in unified text")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[{"a":1}]`, `[{"a":1}]`},
		{"```json\n[{\"a\":1}]\n```", `[{"a":1}]`},
		{"Here is the result:\n[{\"a\":1}]", `[{"a":1}]`},
	}
	for _, tt := range tests {
		got := extractJSON(tt.input)
		if got != tt.expected {
			t.Errorf("extractJSON(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestListToTree_Simple(t *testing.T) {
	flat := []FlatEntry{
		{Structure: "1", Title: "Chapter 1", PhysicalIndex: 1},
		{Structure: "1.1", Title: "Section 1.1", PhysicalIndex: 2},
		{Structure: "1.2", Title: "Section 1.2", PhysicalIndex: 4},
		{Structure: "2", Title: "Chapter 2", PhysicalIndex: 6},
	}
	assignIndices(flat, 10)
	tree := listToTree(flat)

	if len(tree) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(tree))
	}
	if tree[0].Title != "Chapter 1" {
		t.Fatalf("expected Chapter 1, got %s", tree[0].Title)
	}
	if len(tree[0].Nodes) != 2 {
		t.Fatalf("expected 2 children for Chapter 1, got %d", len(tree[0].Nodes))
	}
}

func TestAssignNodeIDs(t *testing.T) {
	tree := []Node{
		{Title: "A", Nodes: []Node{
			{Title: "A.1", Nodes: []Node{}},
			{Title: "A.2", Nodes: []Node{}},
		}},
		{Title: "B", Nodes: []Node{}},
	}

	count := assignNodeIDs(tree, 1)
	if count != 5 { // A=1, A.1=2, A.2=3, B=4, next=5
		t.Fatalf("expected next id 5, got %d", count)
	}
	if tree[0].NodeID != "0001" {
		t.Fatalf("expected 0001, got %s", tree[0].NodeID)
	}
	if tree[0].Nodes[1].NodeID != "0003" {
		t.Fatalf("expected 0003, got %s", tree[0].Nodes[1].NodeID)
	}
	if tree[1].NodeID != "0004" {
		t.Fatalf("expected 0004, got %s", tree[1].NodeID)
	}
}

func TestBuildFlatFallback(t *testing.T) {
	pages := []ExtractionPage{
		{PageNumber: 1}, {PageNumber: 2}, {PageNumber: 3},
	}
	flat := buildFlatFallback(pages)
	if len(flat) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(flat))
	}
	if flat[0].Structure != "1" || flat[2].Structure != "3" {
		t.Fatal("wrong structure codes")
	}
}

func TestParentStructure(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"1", ""},
		{"1.1", "1"},
		{"1.2.3", "1.2"},
		{"1.2.3.4", "1.2.3"},
	}
	for _, tt := range tests {
		got := parentStructure(tt.input)
		if got != tt.expected {
			t.Errorf("parentStructure(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNodeJSON_Roundtrip(t *testing.T) {
	tree := []Node{
		{
			Title:      "Sistema de frenos",
			NodeID:     "0001",
			StartIndex: 34,
			EndIndex:   38,
			Summary:    "Brake system specs",
			Nodes: []Node{
				{
					Title:      "Disco delantero",
					NodeID:     "0002",
					StartIndex: 34,
					EndIndex:   36,
					Summary:    "Front disc specs",
					Nodes:      []Node{},
				},
			},
		},
	}

	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatal(err)
	}

	var parsed []Node
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed[0].Nodes[0].NodeID != "0002" {
		t.Fatal("roundtrip failed")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
