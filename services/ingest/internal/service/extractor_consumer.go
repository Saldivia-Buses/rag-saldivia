package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/Camionerou/rag-saldivia/services/ingest/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/ingest/internal/tree"
)

// ExtractorConsumer subscribes to extractor results and stores pages + generates trees.
type ExtractorConsumer struct {
	nc       *nats.Conn
	pool     *pgxpool.Pool
	repo     *repository.Queries
	treeGen  *tree.Generator
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewExtractorConsumer creates a consumer for extractor results.
func NewExtractorConsumer(nc *nats.Conn, pool *pgxpool.Pool, treeGen *tree.Generator) *ExtractorConsumer {
	return &ExtractorConsumer{
		nc:      nc,
		pool:    pool,
		repo:    repository.New(pool),
		treeGen: treeGen,
	}
}

// ExtractionResult matches the output schema from the Extractor service.
type ExtractionResult struct {
	DocumentID string           `json:"document_id"`
	FileName   string           `json:"file_name"`
	TotalPages int              `json:"total_pages"`
	Pages      []ExtractionPage `json:"pages"`
	Metadata   json.RawMessage  `json:"metadata"`
}

// ExtractionPage is a page from the extractor output.
type ExtractionPage struct {
	PageNumber int             `json:"page_number"`
	Text       string          `json:"text"`
	Tables     json.RawMessage `json:"tables"`
	Images     json.RawMessage `json:"images"`
}

// Start begins consuming extractor results.
func (c *ExtractorConsumer) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	js, err := jetstream.New(c.nc)
	if err != nil {
		return err
	}

	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "EXTRACTOR_RESULTS",
		Subjects: []string{"tenant.*.extractor.result.>"},
	})
	if err != nil {
		return err
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "ingest-extractor-consumer",
		FilterSubject: "tenant.*.extractor.result.>",
		MaxDeliver:    3,
		AckWait:       300_000_000_000, // 5 min for tree generation
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			msg, err := cons.Next(jetstream.FetchMaxWait(30_000_000_000)) // 30s poll
			if err != nil {
				if c.ctx.Err() != nil {
					return // shutting down
				}
				continue
			}
			c.handleResult(msg)
		}
	}()

	slog.Info("extractor result consumer started")
	return nil
}

func (c *ExtractorConsumer) handleResult(msg jetstream.Msg) {
	var result ExtractionResult
	if err := json.Unmarshal(msg.Data(), &result); err != nil {
		slog.Error("invalid extraction result", "error", err)
		msg.Ack()
		return
	}

	slog.Info("received extraction result", "doc_id", result.DocumentID, "pages", result.TotalPages)

	ctx := c.ctx

	// 1. Update document status + page count
	c.repo.UpdateDocumentStatus(ctx, repository.UpdateDocumentStatusParams{
		Status: "indexing",
		ID:     result.DocumentID,
	})
	c.repo.UpdateDocumentPages(ctx, repository.UpdateDocumentPagesParams{
		TotalPages: pgtype.Int4{Int32: int32(result.TotalPages), Valid: true},
		ID:         result.DocumentID,
	})

	// 2. Store pages in document_pages
	for _, page := range result.Pages {
		tables := page.Tables
		if tables == nil {
			tables = json.RawMessage("[]")
		}
		images := page.Images
		if images == nil {
			images = json.RawMessage("[]")
		}
		c.repo.InsertDocumentPage(ctx, repository.InsertDocumentPageParams{
			DocumentID: result.DocumentID,
			PageNumber: int32(page.PageNumber),
			Text:       page.Text,
			Tables:     tables,
			Images:     images,
		})
	}

	// 3. Generate tree (if tree generator is available)
	if c.treeGen != nil {
		treePages := make([]tree.ExtractionPage, len(result.Pages))
		for i, p := range result.Pages {
			treePages[i] = tree.ExtractionPage{
				PageNumber: p.PageNumber,
				Text:       p.Text,
			}
		}

		treeResult, err := c.treeGen.Generate(ctx, treePages, tree.Prompts{})
		if err != nil {
			slog.Error("tree generation failed", "doc_id", result.DocumentID, "error", err)
		} else {
			treeJSON, _ := json.Marshal(treeResult.Tree)
			c.repo.InsertDocumentTree(ctx, repository.InsertDocumentTreeParams{
				DocumentID:     result.DocumentID,
				Tree:           treeJSON,
				DocDescription: pgtype.Text{String: treeResult.DocDescription, Valid: treeResult.DocDescription != ""},
				ModelUsed:      "unknown", // TODO: from config
				NodeCount:      int32(treeResult.NodeCount),
			})
		}
	}

	// 4. Update document status to ready
	c.repo.UpdateDocumentStatus(ctx, repository.UpdateDocumentStatusParams{
		Status: "ready",
		ID:     result.DocumentID,
	})

	msg.Ack()
	slog.Info("extraction result processed", "doc_id", result.DocumentID)
}

// Stop cancels the consumer.
func (c *ExtractorConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}
