package intelligence

import (
	"context"
	"log/slog"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// Engine is the intelligence layer orchestrator. Thread-safe, created once at startup.
type Engine struct {
	registry *DomainRegistry
	log      *slog.Logger
}

// NewEngine creates the intelligence engine with the default domain registry.
func NewEngine(log *slog.Logger) (*Engine, error) {
	domains := defaultDomains()
	registry, err := NewDomainRegistry(domains)
	if err != nil {
		return nil, err
	}
	return &Engine{registry: registry, log: log}, nil
}

// AnalysisRequest holds everything the intelligence layer needs.
type AnalysisRequest struct {
	Query   string
	FullCtx *astrocontext.FullContext
}

// AnalysisResult is the output of the intelligence layer.
type AnalysisResult struct {
	Intent       *Intent
	Domain       *ResolvedDomain
	Gate         *GateResult
	CrossRefs    []CrossReference
	Brief        string // domain-aware intelligence brief
	SystemPrompt string // domain-aware system prompt for LLM
	Warnings     []string
}

// Analyze runs the full intelligence pipeline.
// This is the single entry point for the handler.
func (e *Engine) Analyze(ctx context.Context, req *AnalysisRequest) (*AnalysisResult, error) {
	result := &AnalysisResult{}

	// Step 1: Parse intent from query
	result.Intent = ParseIntent(req.Query, e.registry)
	e.log.Debug("intent parsed",
		"domain", result.Intent.PrimaryDomain,
		"confidence", result.Intent.Confidence,
		"keywords", result.Intent.MatchedKeywords,
	)

	// Step 2: Resolve domain with inheritance
	domain, err := e.registry.Resolve(result.Intent.PrimaryDomain)
	if err != nil {
		// Fallback to predictivo
		domain, _ = e.registry.Resolve("predictivo")
		result.Warnings = append(result.Warnings, "domain resolution failed, using predictivo")
	}
	result.Domain = domain

	// Step 3: Validate technique data richness
	result.Gate = ValidateTechniques(req.FullCtx, domain)
	if result.Gate.Coverage < 0.5 {
		result.Warnings = append(result.Warnings,
			"menos del 50% de técnicas requeridas tienen datos")
	}

	// Step 4: Find cross-references (deterministic)
	result.CrossRefs = AnalyzeCrossReferences(req.FullCtx)
	e.log.Debug("cross-references found", "count", len(result.CrossRefs))

	// Step 5: Build domain-aware intelligence brief
	result.Brief = BuildIntelligenceBrief(req.FullCtx, domain, result.Gate, result.CrossRefs)

	// Step 6: Build domain-aware system prompt
	result.SystemPrompt = BuildSystemPrompt(domain, result.Gate, result.CrossRefs)

	return result, nil
}

// Registry returns the domain registry (for intent parsing in handlers).
func (e *Engine) Registry() *DomainRegistry { return e.registry }
