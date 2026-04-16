package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Purchasing handles purchasing business logic.
type Purchasing struct {
	repo      *repository.Queries
	pool      TxStarter
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewPurchasing creates a purchasing service.
func NewPurchasing(repo *repository.Queries, pool TxStarter, auditWriter *audit.Writer, publisher *traces.Publisher) *Purchasing {
	return &Purchasing{repo: repo, pool: pool, audit: auditWriter, publisher: publisher}
}

func (s *Purchasing) ListOrders(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListPurchaseOrdersRow, error) {
	return s.repo.ListPurchaseOrders(ctx, repository.ListPurchaseOrdersParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

// OrderDetail bundles an order with its lines.
type OrderDetail struct {
	Order repository.ErpPurchaseOrder            `json:"order"`
	Lines []repository.ListPurchaseOrderLinesRow `json:"lines"`
}

func (s *Purchasing) GetOrder(ctx context.Context, id pgtype.UUID, tenantID string) (*OrderDetail, error) {
	order, err := s.repo.GetPurchaseOrder(ctx, repository.GetPurchaseOrderParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	lines, err := s.repo.ListPurchaseOrderLines(ctx, repository.ListPurchaseOrderLinesParams{
		OrderID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list lines: %w", err)
	}
	return &OrderDetail{Order: order, Lines: lines}, nil
}

// CreateOrderRequest holds data for creating a purchase order with lines.
type CreateOrderRequest struct {
	TenantID   string
	Number     string
	Date       pgtype.Date
	SupplierID pgtype.UUID
	CurrencyID pgtype.UUID
	Notes      string
	UserID     string
	IP         string
	Lines      []CreateOrderLineRequest
}

type CreateOrderLineRequest struct {
	ArticleID pgtype.UUID
	Quantity  string
	UnitPrice string
}

func (s *Purchasing) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderDetail, error) {
	if req.Number == "" || len(req.Lines) == 0 {
		return nil, fmt.Errorf("number and at least one line required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	// Calculate total
	var total pgtype.Numeric
	_ = total.Scan("0")

	order, err := qtx.CreatePurchaseOrder(ctx, repository.CreatePurchaseOrderParams{
		TenantID:   req.TenantID,
		Number:     req.Number,
		Date:       req.Date,
		SupplierID: req.SupplierID,
		CurrencyID: req.CurrencyID,
		Total:      total,
		Notes:      req.Notes,
		UserID:     req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	for i, l := range req.Lines {
		_, err := qtx.CreatePurchaseOrderLine(ctx, repository.CreatePurchaseOrderLineParams{
			TenantID:  req.TenantID,
			OrderID:   order.ID,
			ArticleID: l.ArticleID,
			Quantity:  pgNumeric(l.Quantity),
			UnitPrice: pgNumeric(l.UnitPrice),
			SortOrder: int32(i),
		})
		if err != nil {
			return nil, fmt.Errorf("create line %d: %w", i, err)
		}
	}

	// Calculate total from lines (SUM(qty * price)) — update order in same TX
	// Simple sum since pgtype.Numeric arithmetic is complex, use SQL
	_, _ = tx.Exec(ctx,
		`UPDATE erp_purchase_orders SET total = (
			SELECT COALESCE(SUM(quantity * unit_price), 0)
			FROM erp_purchase_order_lines WHERE order_id = $1 AND tenant_id = $2
		) WHERE id = $1 AND tenant_id = $2`,
		order.ID, req.TenantID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit order: %w", err)
	}

	// Re-fetch order with updated total
	order, _ = s.repo.GetPurchaseOrder(ctx, repository.GetPurchaseOrderParams{
		ID: order.ID, TenantID: req.TenantID,
	})

	idStr := uuidStr(order.ID)
	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.purchase_order.created", Resource: idStr,
		Details: map[string]any{"number": req.Number, "lines": len(req.Lines)}, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_purchasing", map[string]any{
		"action": "order_created", "order_id": idStr,
	})

	// Fetch with JOINed data
	lines, _ := s.repo.ListPurchaseOrderLines(ctx, repository.ListPurchaseOrderLinesParams{
		OrderID: order.ID, TenantID: req.TenantID,
	})

	slog.Info("purchase order created", "id", idStr, "number", req.Number)
	return &OrderDetail{Order: order, Lines: lines}, nil
}

func (s *Purchasing) ApproveOrder(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	rows, err := s.repo.ApprovePurchaseOrder(ctx, repository.ApprovePurchaseOrderParams{
		ID: id, TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("approve order: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found or not in draft status")
	}
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.purchase_order.approved", Resource: uuidStr(id), IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_purchasing", map[string]any{
		"action": "order_approved", "order_id": uuidStr(id),
	})
	return nil
}

func (s *Purchasing) ListReceipts(ctx context.Context, tenantID string, limit, offset int) ([]repository.ListPurchaseReceiptsRow, error) {
	return s.repo.ListPurchaseReceipts(ctx, repository.ListPurchaseReceiptsParams{
		TenantID: tenantID, Limit: int32(limit), Offset: int32(offset),
	})
}

// ReceiveRequest holds data for receiving goods against a PO.
type ReceiveRequest struct {
	TenantID string
	OrderID  pgtype.UUID
	Date     pgtype.Date
	Number   string
	UserID   string
	Notes    string
	IP       string
	Lines    []ReceiveLineRequest
}

type ReceiveLineRequest struct {
	OrderLineID pgtype.UUID
	ArticleID   pgtype.UUID
	Quantity    string
}

// ============================================================
// QC Inspection (Plan 18 Fase 3)
// ============================================================

// InspectionInput holds data for one inspection line.
type InspectionInput struct {
	ReceiptLineID string `json:"receipt_line_id"`
	ArticleID     string `json:"article_id"`
	WarehouseID   string `json:"warehouse_id"`
	Quantity      string `json:"quantity"`
	AcceptedQty   string `json:"accepted_qty"`
	RejectedQty   string `json:"rejected_qty"`
	Notes         string `json:"notes"`
}

// InspectReceipt performs QC inspection on receipt lines: accept/reject per line,
// creates stock movements for accepted qty, demerits for rejected qty.
func (s *Purchasing) InspectReceipt(ctx context.Context, tenantID string, receiptID pgtype.UUID, inspections []InspectionInput, inspectorID, ip string) ([]repository.ErpQcInspection, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	// Resolve supplier through receipt → order → supplier_id
	receipt, err := qtx.GetPurchaseReceipt(ctx, repository.GetPurchaseReceiptParams{
		ID: receiptID, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("get receipt: %w", err)
	}
	order, err := qtx.GetPurchaseOrder(ctx, repository.GetPurchaseOrderParams{
		ID: receipt.OrderID, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	supplierID := order.SupplierID

	var results []repository.ErpQcInspection
	for i, inp := range inspections {
		rlID, err := parseUUIDStr(inp.ReceiptLineID)
		if err != nil {
			return nil, fmt.Errorf("invalid receipt_line_id in inspection %d: %w", i, err)
		}
		artID, err := parseUUIDStr(inp.ArticleID)
		if err != nil {
			return nil, fmt.Errorf("invalid article_id in inspection %d: %w", i, err)
		}

		qc, err := qtx.CreateInspection(ctx, repository.CreateInspectionParams{
			TenantID:      tenantID,
			ReceiptID:     receiptID,
			ReceiptLineID: rlID,
			ArticleID:     artID,
			Quantity:      pgNumeric(inp.Quantity),
			AcceptedQty:   pgNumeric(inp.AcceptedQty),
			RejectedQty:   pgNumeric(inp.RejectedQty),
			Status:        "pending",
			InspectorID:   inspectorID,
			Notes:         inp.Notes,
		})
		if err != nil {
			return nil, fmt.Errorf("create inspection %d: %w", i, err)
		}

		// Mark as completed (sets completed_at timestamp). The returned row has
		// the updated status/completed_at so callers observe the final state.
		completed, err := qtx.CompleteInspection(ctx, repository.CompleteInspectionParams{
			ID: qc.ID, TenantID: tenantID,
		})
		if err != nil {
			return nil, fmt.Errorf("complete inspection %d: %w", i, err)
		}
		qc = completed

		// Accepted qty → stock movement (in)
		acceptedF := parseFloatStr(inp.AcceptedQty)
		if acceptedF > 0 {
			whID, err := parseUUIDStr(inp.WarehouseID)
			if err != nil {
				return nil, fmt.Errorf("warehouse_id required for accepted goods in inspection %d: %w", i, err)
			}
			_, err = qtx.CreateStockMovement(ctx, repository.CreateStockMovementParams{
				TenantID:      tenantID,
				ArticleID:     artID,
				WarehouseID:   whID,
				MovementType:  "in",
				Quantity:      pgNumeric(inp.AcceptedQty),
				UnitCost:      pgNumeric("0"),
				ReferenceType: pgText("qc_inspection"),
				ReferenceID:   qc.ID,
				UserID:        inspectorID,
				Notes:         fmt.Sprintf("QC accepted: %s of %s", inp.AcceptedQty, inp.Quantity),
			})
			if err != nil {
				return nil, fmt.Errorf("stock movement for QC accept %d: %w", i, err)
			}
			// Update stock level
			_ = qtx.UpsertStockLevel(ctx, repository.UpsertStockLevelParams{
				TenantID: tenantID, ArticleID: artID,
				WarehouseID: whID, Quantity: pgNumeric(inp.AcceptedQty),
			})
		}

		// Rejected qty → demerit to supplier
		rejectedF := parseFloatStr(inp.RejectedQty)
		if rejectedF > 0 {
			points := int32(rejectedF) // 1 point per rejected unit
			if points < 1 {
				points = 1
			}
			_, err := qtx.CreateDemerit(ctx, repository.CreateDemeritParams{
				TenantID:     tenantID,
				SupplierID:   supplierID,
				InspectionID: qc.ID,
				Points:       points,
				Reason:       fmt.Sprintf("Rechazo QC: %s unidades de %s", inp.RejectedQty, inp.Quantity),
			})
			if err != nil {
				return nil, fmt.Errorf("create demerit %d: %w", i, err)
			}
		}

		results = append(results, qc)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit inspections: %w", err)
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: inspectorID,
		Action: "erp.qc.inspected", Resource: uuidStr(receiptID),
		Details: map[string]any{"inspections": len(inspections)}, IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_purchasing", map[string]any{
		"action": "qc_inspected", "receipt_id": uuidStr(receiptID),
	})

	return results, nil
}

func (s *Purchasing) ListInspections(ctx context.Context, tenantID, status string, limit, offset int) ([]repository.ListInspectionsRow, error) {
	return s.repo.ListInspections(ctx, repository.ListInspectionsParams{
		TenantID: tenantID, StatusFilter: status, Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *Purchasing) GetInspection(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpQcInspection, error) {
	return s.repo.GetInspection(ctx, repository.GetInspectionParams{ID: id, TenantID: tenantID})
}

func (s *Purchasing) ListSupplierDemerits(ctx context.Context, tenantID string, supplierID pgtype.UUID) ([]repository.ErpSupplierDemerit, error) {
	return s.repo.ListSupplierDemerits(ctx, repository.ListSupplierDemeritsParams{
		TenantID: tenantID, SupplierID: supplierID,
	})
}

func (s *Purchasing) GetSupplierDemeritTotal(ctx context.Context, tenantID string, supplierID pgtype.UUID) (int32, error) {
	return s.repo.GetSupplierDemeritTotal(ctx, repository.GetSupplierDemeritTotalParams{
		TenantID: tenantID, SupplierID: supplierID,
	})
}

// parseUUIDStr parses a UUID string for service layer use.
func parseUUIDStr(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, fmt.Errorf("empty UUID")
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// parseFloatStr is an alias for parseFloatApprox (defined in treasury.go).
// Used only for approximate comparisons, never for financial arithmetic.
func parseFloatStr(s string) float64 { return parseFloatApprox(s) }

// Receive creates a receipt and updates received quantities in one transaction.
// Order must be in 'approved' or 'partial' status.
func (s *Purchasing) Receive(ctx context.Context, req ReceiveRequest) error {
	if len(req.Lines) == 0 {
		return fmt.Errorf("at least one line required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.repo.WithTx(tx)

	// Verify order is receivable
	order, err := qtx.GetPurchaseOrder(ctx, repository.GetPurchaseOrderParams{
		ID: req.OrderID, TenantID: req.TenantID,
	})
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}
	if order.Status != "approved" && order.Status != "partial" {
		return fmt.Errorf("order must be approved or partial to receive, current: %s", order.Status)
	}

	receipt, err := qtx.CreatePurchaseReceipt(ctx, repository.CreatePurchaseReceiptParams{
		TenantID: req.TenantID, OrderID: req.OrderID, Date: req.Date,
		Number: req.Number, UserID: req.UserID, Notes: req.Notes,
	})
	if err != nil {
		return fmt.Errorf("create receipt: %w", err)
	}

	for i, l := range req.Lines {
		_, err := qtx.CreatePurchaseReceiptLine(ctx, repository.CreatePurchaseReceiptLineParams{
			TenantID:    req.TenantID,
			ReceiptID:   receipt.ID,
			OrderLineID: l.OrderLineID,
			ArticleID:   l.ArticleID,
			Quantity:    pgNumeric(l.Quantity),
		})
		if err != nil {
			return fmt.Errorf("create receipt line %d: %w", i, err)
		}

		// Update received qty on PO line
		if err := qtx.UpdateReceivedQty(ctx, repository.UpdateReceivedQtyParams{
			ID: l.OrderLineID, TenantID: req.TenantID, ReceivedQty: pgNumeric(l.Quantity),
		}); err != nil {
			return fmt.Errorf("update received qty %d: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit receipt: %w", err)
	}

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID, UserID: req.UserID,
		Action: "erp.purchase.received", Resource: uuidStr(receipt.ID),
		Details: map[string]any{"order_id": uuidStr(req.OrderID), "lines": len(req.Lines)}, IP: req.IP,
	})
	s.publisher.Broadcast(req.TenantID, "erp_purchasing", map[string]any{
		"action": "receipt_created", "receipt_id": uuidStr(receipt.ID),
	})
	return nil
}
