// Package tree implements PageIndex-inspired tree generation from extracted documents.
package tree

// Node represents a single node in the document tree (TOC structure).
// This is the JSON structure stored in document_trees.tree.
type Node struct {
	Title      string `json:"title"`
	NodeID     string `json:"node_id"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Summary    string `json:"summary"`
	Nodes      []Node `json:"nodes"`
}

// FlatEntry is a section detected during TOC/structure extraction,
// before conversion to a nested tree.
type FlatEntry struct {
	Structure     string `json:"structure"`      // "1", "1.1", "1.2.1"
	Title         string `json:"title"`
	PhysicalIndex int    `json:"physical_index"` // 1-based page number
}

// ExtractionPage represents a page from the Extractor output.
type ExtractionPage struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
	Tables     []struct {
		Markdown string `json:"markdown"`
		Caption  string `json:"caption"`
	} `json:"tables"`
	Images []struct {
		Description   string   `json:"description"`
		Type          string   `json:"type"`
		ExtractedData []string `json:"extracted_data"`
	} `json:"images"`
}
