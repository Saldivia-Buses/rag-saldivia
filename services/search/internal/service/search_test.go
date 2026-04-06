package service

import (
	"strings"
	"testing"
)

func TestBuildTreeView_SingleLevel(t *testing.T) {
	var sb strings.Builder
	index := make(map[string]selectedNode)

	nodes := []TreeNode{
		{NodeID: "n1", Title: "Introduction", Summary: "Overview of the doc", StartIndex: 0, EndIndex: 2},
		{NodeID: "n2", Title: "Methods", Summary: "Research methods", StartIndex: 3, EndIndex: 5},
	}

	buildTreeView(&sb, index, nodes, "doc-1", "paper.pdf", "")

	// Verify index populated
	if len(index) != 2 {
		t.Fatalf("expected 2 nodes in index, got %d", len(index))
	}

	n1 := index["n1"]
	if n1.DocumentID != "doc-1" || n1.Title != "Introduction" || n1.StartIndex != 0 || n1.EndIndex != 2 {
		t.Errorf("n1 mismatch: %+v", n1)
	}

	n2 := index["n2"]
	if n2.DocumentName != "paper.pdf" || n2.StartIndex != 3 {
		t.Errorf("n2 mismatch: %+v", n2)
	}

	// Verify output format
	output := sb.String()
	if !strings.Contains(output, "[n1] Introduction") {
		t.Errorf("expected node n1 in output, got: %s", output)
	}
	if !strings.Contains(output, "(pages 3-5)") {
		t.Errorf("expected page range in output, got: %s", output)
	}
}

func TestBuildTreeView_NestedNodes(t *testing.T) {
	var sb strings.Builder
	index := make(map[string]selectedNode)

	nodes := []TreeNode{
		{
			NodeID: "ch1", Title: "Chapter 1", Summary: "First chapter", StartIndex: 0, EndIndex: 10,
			Nodes: []TreeNode{
				{NodeID: "s1.1", Title: "Section 1.1", Summary: "First section", StartIndex: 0, EndIndex: 3},
				{NodeID: "s1.2", Title: "Section 1.2", Summary: "Second section", StartIndex: 4, EndIndex: 10},
			},
		},
	}

	buildTreeView(&sb, index, nodes, "doc-1", "book.pdf", "")

	if len(index) != 3 {
		t.Fatalf("expected 3 nodes (parent + 2 children), got %d", len(index))
	}

	// Children should have indentation in output
	output := sb.String()
	if !strings.Contains(output, "  [s1.1]") {
		t.Errorf("expected indented child node, got: %s", output)
	}
}

func TestParseNodeIDs_CommaSeparated(t *testing.T) {
	// Simulate what navigateTrees does with LLM response
	tests := []struct {
		name     string
		response string
		nodeIDs  map[string]selectedNode
		want     int
	}{
		{
			name:     "normal response",
			response: "n1, n2, n3",
			nodeIDs:  map[string]selectedNode{"n1": {NodeID: "n1"}, "n2": {NodeID: "n2"}, "n3": {NodeID: "n3"}},
			want:     3,
		},
		{
			name:     "with extra whitespace",
			response: "  n1 , n2  ",
			nodeIDs:  map[string]selectedNode{"n1": {NodeID: "n1"}, "n2": {NodeID: "n2"}},
			want:     2,
		},
		{
			name:     "some invalid IDs filtered out",
			response: "n1, invalid, n2",
			nodeIDs:  map[string]selectedNode{"n1": {NodeID: "n1"}, "n2": {NodeID: "n2"}},
			want:     2,
		},
		{
			name:     "all invalid",
			response: "foo, bar, baz",
			nodeIDs:  map[string]selectedNode{"n1": {NodeID: "n1"}},
			want:     0,
		},
		{
			name:     "empty response",
			response: "",
			nodeIDs:  map[string]selectedNode{"n1": {NodeID: "n1"}},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := strings.Split(strings.TrimSpace(tt.response), ",")
			var selected []selectedNode
			for _, raw := range ids {
				id := strings.TrimSpace(raw)
				if node, ok := tt.nodeIDs[id]; ok {
					selected = append(selected, node)
				}
			}
			if len(selected) != tt.want {
				t.Errorf("expected %d selected nodes, got %d", tt.want, len(selected))
			}
		})
	}
}

func TestContainsInt(t *testing.T) {
	tests := []struct {
		slice []int
		val   int
		want  bool
	}{
		{[]int{1, 2, 3}, 2, true},
		{[]int{1, 2, 3}, 4, false},
		{[]int{}, 1, false},
		{nil, 1, false},
	}

	for _, tt := range tests {
		got := containsInt(tt.slice, tt.val)
		if got != tt.want {
			t.Errorf("containsInt(%v, %d) = %v, want %v", tt.slice, tt.val, got, tt.want)
		}
	}
}
