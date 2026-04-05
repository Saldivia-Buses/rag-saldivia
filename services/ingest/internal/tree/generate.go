package tree

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/Camionerou/rag-saldivia/services/ingest/internal/llm"
)

// Generator builds a document tree from extracted pages using LLM calls.
type Generator struct {
	llmClient *llm.Client
}

// NewGenerator creates a tree generator.
func NewGenerator(llmClient *llm.Client) *Generator {
	return &Generator{llmClient: llmClient}
}

// GenerateResult is the output of tree generation.
type GenerateResult struct {
	Tree           []Node
	DocDescription string
	NodeCount      int
}

// Generate runs the full tree generation pipeline on extracted pages.
func (g *Generator) Generate(ctx context.Context, pages []ExtractionPage, prompts Prompts) (*GenerateResult, error) {
	totalPages := len(pages)
	if totalPages == 0 {
		return nil, fmt.Errorf("no pages to process")
	}

	// 1. Build unified page-indexed text
	unified := buildUnifiedText(pages)

	// 2. Generate structure via LLM
	flat, err := g.generateStructure(ctx, unified, totalPages, prompts.TreeGeneration)
	if err != nil {
		slog.Warn("tree generation failed, falling back to flat tree", "error", err)
		flat = buildFlatFallback(pages)
	}

	// 3. Assign start/end indices
	assignIndices(flat, totalPages)

	// 4. Convert to nested tree
	tree := listToTree(flat)

	// 5. Assign sequential node IDs
	nodeCount := assignNodeIDs(tree, 1)

	// 6. Generate summaries in parallel
	g.generateSummaries(ctx, tree, pages, prompts.TreeSummary)

	// 7. Generate doc description
	docDesc, _ := g.generateDocDescription(ctx, tree, prompts.DocDescription)

	return &GenerateResult{
		Tree:           tree,
		DocDescription: docDesc,
		NodeCount:      nodeCount,
	}, nil
}

// Prompts holds the prompt templates for each step. Loaded from prompt_versions.
type Prompts struct {
	TreeGeneration string
	TreeSummary    string
	DocDescription string
}

// buildUnifiedText creates a page-indexed text with markers.
func buildUnifiedText(pages []ExtractionPage) string {
	var sb strings.Builder
	for _, p := range pages {
		fmt.Fprintf(&sb, "<page_%d>\n", p.PageNumber)
		sb.WriteString(p.Text)
		sb.WriteString("\n")
		for _, t := range p.Tables {
			sb.WriteString(t.Markdown)
			sb.WriteString("\n")
		}
		for _, img := range p.Images {
			fmt.Fprintf(&sb, "[IMAGEN: %s]\n", img.Description)
		}
		fmt.Fprintf(&sb, "</page_%d>\n\n", p.PageNumber)
	}
	return sb.String()
}

// generateStructure asks the LLM to produce a flat list of sections.
func (g *Generator) generateStructure(ctx context.Context, text string, totalPages int, prompt string) ([]FlatEntry, error) {
	// For large documents, process in chunks
	const maxChunkChars = 30000

	if len(text) <= maxChunkChars {
		return g.generateStructureChunk(ctx, text, prompt)
	}

	// Split into chunks by page boundaries
	chunks := splitByPages(text, maxChunkChars)
	var allEntries []FlatEntry

	for i, chunk := range chunks {
		var contextMsg string
		if i > 0 && len(allEntries) > 0 {
			prev, _ := json.Marshal(allEntries)
			contextMsg = fmt.Sprintf("Previous structure so far:\n%s\n\nContinue with the next part:\n", string(prev))
		}
		entries, err := g.generateStructureChunk(ctx, contextMsg+chunk, prompt)
		if err != nil {
			return nil, fmt.Errorf("chunk %d: %w", i, err)
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

func (g *Generator) generateStructureChunk(ctx context.Context, text, prompt string) ([]FlatEntry, error) {
	fullPrompt := prompt + "\n\nDocument text:\n" + text

	content, err := g.llmClient.SimplePrompt(ctx, fullPrompt, 0.0)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}

	// Parse JSON response — try to extract JSON array from response
	content = extractJSON(content)

	var entries []FlatEntry
	if err := json.Unmarshal([]byte(content), &entries); err != nil {
		return nil, fmt.Errorf("parse structure JSON: %w (raw: %.200s)", err, content)
	}

	return entries, nil
}

// extractJSON strips markdown fences and finds the JSON array.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip markdown code fences
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	// Find first [ and last ]
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// splitByPages splits unified text into chunks at page boundaries.
func splitByPages(text string, maxChars int) []string {
	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxChars {
			chunks = append(chunks, text)
			break
		}
		// Find a page boundary near maxChars
		cutoff := maxChars
		idx := strings.LastIndex(text[:cutoff], "</page_")
		if idx > 0 {
			// Find the end of this tag
			end := strings.Index(text[idx:], "\n")
			if end > 0 {
				cutoff = idx + end + 1
			}
		}
		chunks = append(chunks, text[:cutoff])
		text = text[cutoff:]
	}
	return chunks
}

// buildFlatFallback creates a simple 1-node-per-page tree when LLM fails.
func buildFlatFallback(pages []ExtractionPage) []FlatEntry {
	entries := make([]FlatEntry, len(pages))
	for i, p := range pages {
		entries[i] = FlatEntry{
			Structure:     fmt.Sprintf("%d", i+1),
			Title:         fmt.Sprintf("Page %d", p.PageNumber),
			PhysicalIndex: p.PageNumber,
		}
	}
	return entries
}

// assignIndices sets start_index and end_index for each entry.
func assignIndices(entries []FlatEntry, totalPages int) {
	for i := range entries {
		entries[i].PhysicalIndex = max(1, entries[i].PhysicalIndex)
	}
	// Sort by physical index
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PhysicalIndex < entries[j].PhysicalIndex
	})
}

// listToTree converts a flat list with structure codes to a nested tree.
func listToTree(flat []FlatEntry) []Node {
	if len(flat) == 0 {
		return nil
	}

	totalPages := flat[len(flat)-1].PhysicalIndex

	// Build pointer-based nodes first, then convert to value tree
	type pNode struct {
		node      Node
		structure string
		children  []*pNode
	}

	pNodes := make([]*pNode, len(flat))
	nodeMap := make(map[string]*pNode)

	for i, e := range flat {
		endIdx := totalPages
		if i < len(flat)-1 {
			endIdx = flat[i+1].PhysicalIndex - 1
			if endIdx < e.PhysicalIndex {
				endIdx = e.PhysicalIndex
			}
		}
		pNodes[i] = &pNode{
			node: Node{
				Title:      e.Title,
				StartIndex: e.PhysicalIndex,
				EndIndex:   endIdx,
			},
			structure: e.Structure,
		}
		nodeMap[e.Structure] = pNodes[i]
	}

	// Link parents to children
	var roots []*pNode
	for _, pn := range pNodes {
		parent := parentStructure(pn.structure)
		if parent == "" {
			roots = append(roots, pn)
		} else if p, ok := nodeMap[parent]; ok {
			p.children = append(p.children, pn)
		} else {
			roots = append(roots, pn)
		}
	}

	// Convert pointer tree to value tree
	var convert func(pns []*pNode) []Node
	convert = func(pns []*pNode) []Node {
		result := make([]Node, len(pns))
		for i, pn := range pns {
			result[i] = pn.node
			if len(pn.children) > 0 {
				result[i].Nodes = convert(pn.children)
			} else {
				result[i].Nodes = []Node{}
			}
		}
		return result
	}

	return convert(roots)
}

// parentStructure returns the parent code: "1.2.3" → "1.2", "1" → ""
func parentStructure(s string) string {
	idx := strings.LastIndex(s, ".")
	if idx < 0 {
		return ""
	}
	return s[:idx]
}

// assignNodeIDs assigns zero-padded sequential IDs to all nodes.
func assignNodeIDs(nodes []Node, startID int) int {
	id := startID
	for i := range nodes {
		nodes[i].NodeID = fmt.Sprintf("%04d", id)
		id++
		id = assignNodeIDs(nodes[i].Nodes, id)
	}
	return id
}

// generateSummaries generates a summary for each leaf/branch node in parallel.
func (g *Generator) generateSummaries(ctx context.Context, nodes []Node, pages []ExtractionPage, prompt string) {
	pageMap := make(map[int]ExtractionPage, len(pages))
	for _, p := range pages {
		pageMap[p.PageNumber] = p
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // max 5 concurrent LLM calls

	var generateForNodes func(nodes []Node)
	generateForNodes = func(nodes []Node) {
		for i := range nodes {
			node := &nodes[i]

			wg.Add(1)
			// B5: context-aware semaphore — don't block forever on cancellation
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				wg.Done()
				return
			}
			go func(n *Node) {
				defer wg.Done()
				defer func() { <-sem }()

				// Build text from node's page range
				var text strings.Builder
				for pg := n.StartIndex; pg <= n.EndIndex; pg++ {
					if p, ok := pageMap[pg]; ok {
						text.WriteString(p.Text)
						text.WriteString("\n")
					}
				}

				pageText := text.String()
				if len(pageText) > 8000 {
					pageText = pageText[:8000]
				}

				summary, err := g.llmClient.SimplePrompt(ctx, prompt+"\n\n"+pageText, 0.0)
				if err != nil {
					slog.Warn("summary generation failed", "node", n.NodeID, "error", err)
					n.Summary = n.Title
					return
				}
				n.Summary = strings.TrimSpace(summary)
			}(node)

			if len(node.Nodes) > 0 {
				generateForNodes(node.Nodes)
			}
		}
	}

	generateForNodes(nodes)
	wg.Wait()
}

// generateDocDescription generates a one-line description of the document.
func (g *Generator) generateDocDescription(ctx context.Context, tree []Node, prompt string) (string, error) {
	// Build a brief overview from top-level titles
	var titles []string
	for _, n := range tree {
		titles = append(titles, n.Title)
	}
	overview := strings.Join(titles, ", ")

	content, err := g.llmClient.SimplePrompt(ctx, prompt+"\n\nDocument sections: "+overview, 0.0)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(content), nil
}
