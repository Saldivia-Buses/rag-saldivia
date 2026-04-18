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

func TestBuildTreeView_EmptyNodes(t *testing.T) {
	var sb strings.Builder
	index := make(map[string]selectedNode)

	buildTreeView(&sb, index, []TreeNode{}, "doc-1", "empty.pdf", "")

	if len(index) != 0 {
		t.Fatalf("expected empty index, got %d entries", len(index))
	}
	if sb.Len() != 0 {
		t.Errorf("expected empty output, got: %q", sb.String())
	}
}

func TestBuildTreeView_DeeplyNested(t *testing.T) {
	// 3 levels deep — verify indentation grows correctly
	var sb strings.Builder
	index := make(map[string]selectedNode)

	nodes := []TreeNode{
		{
			NodeID: "root", Title: "Root", Summary: "root", StartIndex: 0, EndIndex: 100,
			Nodes: []TreeNode{
				{
					NodeID: "child", Title: "Child", Summary: "child", StartIndex: 0, EndIndex: 50,
					Nodes: []TreeNode{
						{NodeID: "leaf", Title: "Leaf", Summary: "leaf", StartIndex: 0, EndIndex: 10},
					},
				},
			},
		},
	}

	buildTreeView(&sb, index, nodes, "doc-1", "deep.pdf", "")

	if len(index) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(index))
	}

	output := sb.String()
	// Root has no indent, child has 2 spaces, leaf has 4 spaces
	if !strings.Contains(output, "[root]") {
		t.Error("expected [root] in output")
	}
	if !strings.Contains(output, "  [child]") {
		t.Error("expected '  [child]' (2 spaces) in output")
	}
	if !strings.Contains(output, "    [leaf]") {
		t.Error("expected '    [leaf]' (4 spaces) in output")
	}
}

func TestBuildTreeView_NodeIDPreservedInIndex(t *testing.T) {
	var sb strings.Builder
	index := make(map[string]selectedNode)

	nodes := []TreeNode{
		{NodeID: "xyz-123", Title: "Test Node", Summary: "Test", StartIndex: 5, EndIndex: 10},
	}

	buildTreeView(&sb, index, nodes, "doc-42", "myfile.pdf", "")

	node, ok := index["xyz-123"]
	if !ok {
		t.Fatal("expected node 'xyz-123' in index")
	}
	if node.NodeID != "xyz-123" {
		t.Errorf("expected NodeID 'xyz-123', got %q", node.NodeID)
	}
	if node.StartIndex != 5 || node.EndIndex != 10 {
		t.Errorf("expected StartIndex=5, EndIndex=10, got %d, %d", node.StartIndex, node.EndIndex)
	}
	if node.DocumentID != "doc-42" {
		t.Errorf("expected DocumentID 'doc-42', got %q", node.DocumentID)
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

// TestContainsInt_SingleElement ensures both found and not-found cases
// with a single-element slice are correct.
func TestContainsInt_SingleElement(t *testing.T) {
	if !containsInt([]int{42}, 42) {
		t.Error("expected true for single-element match")
	}
	if containsInt([]int{42}, 0) {
		t.Error("expected false for single-element non-match")
	}
}

// TestBuildTreeView_MultipleDocuments verifies that the same function
// can be called for multiple documents and all entries go into the shared index.
func TestBuildTreeView_MultipleDocuments(t *testing.T) {
	var sb strings.Builder
	index := make(map[string]selectedNode)

	doc1Nodes := []TreeNode{
		{NodeID: "d1-n1", Title: "Doc1 Chapter", Summary: "First doc", StartIndex: 0, EndIndex: 5},
	}
	doc2Nodes := []TreeNode{
		{NodeID: "d2-n1", Title: "Doc2 Chapter", Summary: "Second doc", StartIndex: 0, EndIndex: 3},
	}

	buildTreeView(&sb, index, doc1Nodes, "doc-1", "file1.pdf", "")
	buildTreeView(&sb, index, doc2Nodes, "doc-2", "file2.pdf", "")

	if len(index) != 2 {
		t.Fatalf("expected 2 nodes across two docs, got %d", len(index))
	}

	if index["d1-n1"].DocumentID != "doc-1" {
		t.Errorf("expected d1-n1 to belong to doc-1")
	}
	if index["d2-n1"].DocumentID != "doc-2" {
		t.Errorf("expected d2-n1 to belong to doc-2")
	}
}
