package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Worker consumes ingest jobs from NATS JetStream and forwards documents
// to the NVIDIA RAG Blueprint for vectorization.
type Worker struct {
	nc        *nats.Conn
	svc       *Ingest
	publisher EventPublisher
	client    *http.Client
	cfg       Config
	cons      jetstream.Consumer
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorker creates an ingest worker.
func NewWorker(nc *nats.Conn, svc *Ingest, publisher EventPublisher, cfg Config) *Worker {
	return &Worker{
		nc:        nc,
		svc:       svc,
		publisher: publisher,
		client:    &http.Client{Timeout: cfg.Timeout},
		cfg:       cfg,
	}
}

const (
	streamName    = "INGEST"
	durableName   = "ingest-worker"
	subjectFilter = "tenant.*.ingest.process"
	maxDeliveries = 3
)

// Start creates a JetStream durable consumer and begins processing.
func (w *Worker) Start(ctx context.Context) error {
	js, err := jetstream.New(w.nc)
	if err != nil {
		return fmt.Errorf("create jetstream context: %w", err)
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectFilter},
		Storage:  jetstream.FileStorage,
		MaxAge:   24 * 60 * 60 * 1e9, // 24h retention
	})
	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	cons, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       durableName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: subjectFilter,
		MaxDeliver:    maxDeliveries,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}
	w.cons = cons
	w.ctx, w.cancel = context.WithCancel(ctx)

	go w.consumeLoop()

	slog.Info("ingest worker started (JetStream durable)", "stream", streamName, "consumer", durableName)
	return nil
}

// Stop cancels the worker loop.
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Worker) consumeLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		batch, err := w.cons.Fetch(1, jetstream.FetchMaxWait(5e9)) // 1 at a time, 5s wait
		if err != nil {
			if w.ctx.Err() != nil {
				return
			}
			slog.Warn("jetstream fetch error", "error", err)
			continue
		}

		for msg := range batch.Messages() {
			w.processJob(msg)
		}
	}
}

func (w *Worker) processJob(msg jetstream.Msg) {
	var im IngestMessage
	if err := json.Unmarshal(msg.Data(), &im); err != nil {
		slog.Warn("invalid ingest message", "error", err, "subject", msg.Subject())
		_ = msg.Term()
		return
	}

	if im.JobID == "" || im.StagedPath == "" {
		slog.Warn("ingest message missing required fields", "subject", msg.Subject())
		_ = msg.Term()
		return
	}

	// S4: validate tenant from NATS subject matches payload
	if subjectTenant := tenantFromSubject(msg.Subject()); subjectTenant != im.TenantSlug {
		slog.Warn("tenant mismatch between subject and payload", "subject", subjectTenant, "payload", im.TenantSlug)
		_ = msg.Term()
		return
	}

	ctx := w.ctx
	slog.Info("processing ingest job", "job_id", im.JobID, "file", im.FileName, "collection", im.Collection)

	// Update status to processing
	_ = w.svc.UpdateJobStatus(ctx, im.JobID, "processing", nil)

	// Forward to Blueprint
	if err := w.forwardToBlueprint(ctx, im); err != nil {
		slog.Error("ingest job failed", "job_id", im.JobID, "error", err)

		// D5: only mark "failed" on final attempt, keep "processing" during retries
		meta, _ := msg.Metadata()
		if meta != nil && meta.NumDelivered >= maxDeliveries {
			errMsg := fmt.Sprintf("failed after %d attempts: %v", maxDeliveries, err)
			_ = w.svc.UpdateJobStatus(ctx, im.JobID, "failed", &errMsg)
			_ = os.Remove(im.StagedPath)
			_ = msg.Term()
		} else {
			_ = msg.Nak() // retry — status stays as "processing"
		}
		return
	}

	// Mark completed and clean up staged file
	_ = w.svc.UpdateJobStatus(ctx, im.JobID, "completed", nil)
	_ = os.Remove(im.StagedPath)

	// Publish notification + WS event
	w.publishCompletion(im)

	_ = msg.Ack()
	slog.Info("ingest job completed", "job_id", im.JobID, "file", im.FileName)
}

func (w *Worker) forwardToBlueprint(ctx context.Context, im IngestMessage) error {
	file, err := os.Open(im.StagedPath)
	if err != nil {
		return fmt.Errorf("open staged file: %w", err)
	}
	defer func() { _ = file.Close() }()

	namespacedCollection := im.TenantSlug + "-" + im.Collection

	// Build multipart request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", im.FileName)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file to form: %w", err)
	}
	if err := writer.WriteField("collection_name", namespacedCollection); err != nil {
		return fmt.Errorf("write collection field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		w.cfg.BlueprintURL+"/v1/documents", &body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("blueprint request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// D3: log full response body but don't expose to clients
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		slog.Error("blueprint error response", "status", resp.StatusCode, "body", string(respBody), "job_id", im.JobID)
		return fmt.Errorf("blueprint returned HTTP %d", resp.StatusCode)
	}

	return nil
}

func (w *Worker) publishCompletion(im IngestMessage) {
	// D4: use actual user_id from the ingest message
	if w.publisher != nil {
		evt := map[string]string{
			"type":    "ingest.completed",
			"user_id": im.UserID,
			"title":   "Documento procesado",
			"body":    im.FileName + " ingresado en " + im.Collection,
		}
		if err := w.publisher.Notify(im.TenantSlug, evt); err != nil {
			slog.Warn("failed to publish ingest notification", "error", err)
		}
	}

	// WS broadcast for real-time progress
	payload, _ := json.Marshal(map[string]any{
		"type":    "event",
		"channel": "ingest.jobs",
		"data": map[string]string{
			"job_id":     im.JobID,
			"status":     "completed",
			"file_name":  im.FileName,
			"collection": im.Collection,
		},
	})
	subject := "tenant." + im.TenantSlug + ".ingest.jobs"
	if err := w.nc.Publish(subject, payload); err != nil {
		slog.Warn("failed to broadcast ingest status", "error", err)
	}
}

// tenantFromSubject extracts the tenant slug from a NATS subject.
// Subject format: tenant.{slug}.ingest.process
func tenantFromSubject(subject string) string {
	parts := strings.SplitN(subject, ".", 4)
	if len(parts) < 3 || parts[0] != "tenant" {
		return ""
	}
	return parts[1]
}
