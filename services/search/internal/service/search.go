// Package service implements the Search Service business logic.
// Tree search in 3 phases: LLM navigation → page extraction → citation assembly.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TreeNode mirrors the tree JSON stored in document_trees.
type TreeNode struct {
	Title      string     `json:"title"`
	NodeID     string     `json:"node_id"`
	StartIndex int        `json:"start_index"`
	EndIndex   int        `json:"end_index"`
	Summary    string     `json:"summary"`
	Nodes      []TreeNode `json:"nodes"`
}

// Selection is one document's contribution to a search result.
type Selection struct {
	Document   string          `json:"document"`
	DocumentID string          `json:"document_id"`
	NodeIDs    []string        `json:"node_ids"`
	Sections   []string        `json:"sections"`
	Pages      []int           `json:"pages"`
	Text       string          `json:"text"`
	Tables     json.RawMessage `json:"tables"`
	Images     json.RawMessage `json:"images"`
}

// SearchResult is the response from a search query.
type SearchResult struct {
	Query      string      `json:"query"`
	Selections []Selection `json:"selections"`
	DurationMS int         `json:"duration_ms"`
}

// Search implements the 3-phase tree search.
type Search struct {
	pool        *pgxpool.Pool
	llmEndpoint string
	llmModel    string
	httpClient  *http.Client
}

// New creates a Search service.
func New(pool *pgxpool.Pool, llmEndpoint, llmModel string) *Search {
	return &Search{
		pool:        pool,
		llmEndpoint: llmEndpoint,
		llmModel:    llmModel,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}
}

// SearchDocuments runs the 3-phase tree search for a query.
func (s *Search) SearchDocuments(ctx context.Context, query string, collectionID string, maxNodes int) (*SearchResult, error) {
	start := time.Now()

	if maxNodes <= 0 {
		maxNodes = 5
	}

	// Load trees from DB
	trees, err := s.loadTrees(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("load trees: %w", err)
	}
	if len(trees) == 0 {
		return &SearchResult{Query: query, Selections: []Selection{}}, nil
	}

	// Phase A: LLM navigates the trees (titles + summaries only)
	selectedNodes, err := s.navigateTrees(ctx, query, trees, maxNodes)
	if err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	// Phase B: Extract pages for selected nodes (pure code, no LLM)
	selections, err := s.extractPages(ctx, selectedNodes)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	return &SearchResult{
		Query:      query,
		Selections: selections,
		DurationMS: int(time.Since(start).Milliseconds()),
	}, nil
}

type docTree struct {
	DocumentID     string     `json:"document_id"`
	DocumentName   string     `json:"document_name"`
	DocDescription string     `json:"doc_description"`
	Tree           []TreeNode `json:"tree"`
}

func (s *Search) loadTrees(ctx context.Context, collectionID string) ([]docTree, error) {
	var query string
	var args []any

	if collectionID != "" {
		query = `SELECT dt.document_id, d.name, dt.doc_description, dt.tree
			FROM document_trees dt
			JOIN documents d ON d.id = dt.document_id
			JOIN collection_documents cd ON cd.document_id = dt.document_id
			WHERE cd.collection_id = $1 AND d.status = 'ready'
			ORDER BY dt.created_at DESC`
		args = []any{collectionID}
	} else {
		query = `SELECT dt.document_id, d.name, dt.doc_description, dt.tree
			FROM document_trees dt
			JOIN documents d ON d.id = dt.document_id
			WHERE d.status = 'ready'
			ORDER BY dt.created_at DESC`
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trees []docTree
	for rows.Next() {
		var dt docTree
		var treeJSON []byte
		var docDesc *string
		if err := rows.Scan(&dt.DocumentID, &dt.DocumentName, &docDesc, &treeJSON); err != nil {
			return nil, err
		}
		if docDesc != nil {
			dt.DocDescription = *docDesc
		}
		if err := json.Unmarshal(treeJSON, &dt.Tree); err != nil {
			slog.Warn("skip tree with invalid JSON", "doc_id", dt.DocumentID, "error", err)
			continue
		}
		trees = append(trees, dt)
	}
	return trees, rows.Err()
}

type selectedNode struct {
	DocumentID   string
	DocumentName string
	NodeID       string
	Title        string
	StartIndex   int
	EndIndex     int
}

// navigateTrees sends the tree TOC to the LLM and gets back node IDs.
func (s *Search) navigateTrees(ctx context.Context, query string, trees []docTree, maxNodes int) ([]selectedNode, error) {
	// Build compact tree view (titles + summaries only, no full text)
	var treeView strings.Builder
	nodeIndex := make(map[string]selectedNode) // nodeID → node info

	for _, dt := range trees {
		fmt.Fprintf(&treeView, "Document: %s (%s)\n", dt.DocumentName, dt.DocDescription)
		buildTreeView(&treeView, nodeIndex, dt.Tree, dt.DocumentID, dt.DocumentName, "")
		treeView.WriteString("\n")
	}

	prompt := fmt.Sprintf(
		"Given this question: %q\n\n"+
			"And these document trees (titles and summaries):\n%s\n"+
			"Select up to %d node IDs that are most likely to contain the answer. "+
			"Return ONLY a comma-separated list of node_ids. Nothing else.",
		query, treeView.String(), maxNodes,
	)

	content, err := s.llmChat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response — expecting comma-separated node IDs
	ids := strings.Split(strings.TrimSpace(content), ",")
	var selected []selectedNode
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if node, ok := nodeIndex[id]; ok {
			selected = append(selected, node)
		}
	}

	if len(selected) == 0 {
		slog.Warn("llm returned no valid node IDs", "raw", content)
	}

	return selected, nil
}

func buildTreeView(sb *strings.Builder, index map[string]selectedNode, nodes []TreeNode, docID, docName, indent string) {
	for _, n := range nodes {
		fmt.Fprintf(sb, "%s[%s] %s — %s (pages %d-%d)\n", indent, n.NodeID, n.Title, n.Summary, n.StartIndex, n.EndIndex)
		index[n.NodeID] = selectedNode{
			DocumentID:   docID,
			DocumentName: docName,
			NodeID:       n.NodeID,
			Title:        n.Title,
			StartIndex:   n.StartIndex,
			EndIndex:     n.EndIndex,
		}
		if len(n.Nodes) > 0 {
			buildTreeView(sb, index, n.Nodes, docID, docName, indent+"  ")
		}
	}
}

// extractPages reads the actual page content for selected nodes from DB.
func (s *Search) extractPages(ctx context.Context, nodes []selectedNode) ([]Selection, error) {
	// Group by document
	byDoc := make(map[string]*Selection)
	docOrder := make([]string, 0)

	for _, n := range nodes {
		sel, ok := byDoc[n.DocumentID]
		if !ok {
			sel = &Selection{
				Document:   n.DocumentName,
				DocumentID: n.DocumentID,
			}
			byDoc[n.DocumentID] = sel
			docOrder = append(docOrder, n.DocumentID)
		}
		sel.NodeIDs = append(sel.NodeIDs, n.NodeID)
		sel.Sections = append(sel.Sections, n.Title)

		// Add page range
		for pg := n.StartIndex; pg <= n.EndIndex; pg++ {
			if !containsInt(sel.Pages, pg) {
				sel.Pages = append(sel.Pages, pg)
			}
		}
	}

	// Fetch pages from DB for each document
	for _, docID := range docOrder {
		sel := byDoc[docID]
		if len(sel.Pages) == 0 {
			continue
		}

		// Query pages in range
		minPage, maxPage := sel.Pages[0], sel.Pages[0]
		for _, p := range sel.Pages {
			if p < minPage {
				minPage = p
			}
			if p > maxPage {
				maxPage = p
			}
		}

		rows, err := s.pool.Query(ctx,
			`SELECT page_number, text, tables, images FROM document_pages
			 WHERE document_id = $1 AND page_number >= $2 AND page_number <= $3
			 ORDER BY page_number`,
			docID, minPage, maxPage,
		)
		if err != nil {
			return nil, fmt.Errorf("query pages doc=%s: %w", docID, err)
		}

		var textParts []string
		var allTables, allImages []json.RawMessage

		for rows.Next() {
			var pageNum int
			var text string
			var tables, images []byte
			if err := rows.Scan(&pageNum, &text, &tables, &images); err != nil {
				rows.Close()
				return nil, err
			}
			textParts = append(textParts, text)
			if len(tables) > 2 { // not "[]"
				allTables = append(allTables, tables)
			}
			if len(images) > 2 {
				allImages = append(allImages, images)
			}
		}
		rows.Close()

		sel.Text = strings.Join(textParts, "\n\n")
		sel.Tables, _ = json.Marshal(allTables)
		sel.Images, _ = json.Marshal(allImages)
	}

	// Preserve order
	result := make([]Selection, 0, len(docOrder))
	for _, docID := range docOrder {
		result = append(result, *byDoc[docID])
	}
	return result, nil
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// llmChat sends a single prompt to the LLM via OpenAI-compatible API.
func (s *Search) llmChat(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": s.llmModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.0,
		"max_tokens":  512,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.llmEndpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return chatResp.Choices[0].Message.Content, nil
}
