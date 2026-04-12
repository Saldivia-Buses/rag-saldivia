-- ═══════════════════════════════════════════════════════════════════════════
-- Plan 20: BI & Analytics — sqlc queries for all reporting endpoints
-- ═══════════════════════════════════════════════════════════════════════════

-- ─── Accounting ────────────────────────────────────────────────────────────

-- name: AnalyticsBalanceEvolution :many
SELECT
    to_char(je.date, 'YYYY-MM') AS period,
    a.account_type,
    SUM(jl.debit) AS total_debit,
    SUM(jl.credit) AS total_credit,
    SUM(jl.debit - jl.credit) AS net
FROM erp_journal_entries je
JOIN erp_journal_lines jl ON jl.entry_id = je.id AND jl.tenant_id = je.tenant_id
JOIN erp_accounts a ON a.id = jl.account_id
WHERE je.tenant_id = $1 AND je.status = 'posted'
    AND je.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(je.date, 'YYYY-MM'), a.account_type
ORDER BY period;

-- name: AnalyticsIncomeExpense :many
SELECT
    to_char(je.date, 'YYYY-MM') AS period,
    SUM(CASE WHEN a.account_type = 'income' THEN jl.credit - jl.debit ELSE 0 END) AS income,
    SUM(CASE WHEN a.account_type = 'expense' THEN jl.debit - jl.credit ELSE 0 END) AS expense
FROM erp_journal_entries je
JOIN erp_journal_lines jl ON jl.entry_id = je.id AND jl.tenant_id = je.tenant_id
JOIN erp_accounts a ON a.id = jl.account_id
WHERE je.tenant_id = $1 AND je.status = 'posted'
    AND je.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(je.date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsCostCenterBreakdown :many
SELECT
    COALESCE(cc.name, 'Sin centro') AS cost_center,
    a.account_type,
    SUM(jl.debit - jl.credit) AS total
FROM erp_journal_lines jl
JOIN erp_journal_entries je ON je.id = jl.entry_id AND je.tenant_id = jl.tenant_id
JOIN erp_accounts a ON a.id = jl.account_id
LEFT JOIN erp_cost_centers cc ON cc.id = jl.cost_center_id
WHERE jl.tenant_id = $1 AND je.status = 'posted'
    AND je.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
    AND a.account_type = 'expense'
GROUP BY cc.name, a.account_type
ORDER BY total DESC;

-- ─── Treasury ──────────────────────────────────────────────────────────────

-- name: AnalyticsCashFlow :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    SUM(CASE WHEN movement_type IN ('cash_in', 'bank_deposit', 'check_received') THEN amount ELSE 0 END) AS inflow,
    SUM(CASE WHEN movement_type IN ('cash_out', 'bank_withdrawal', 'check_issued') THEN amount ELSE 0 END) AS outflow
FROM erp_treasury_movements
WHERE tenant_id = $1 AND status = 'confirmed'
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsTreasuryBalanceEvolution :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    ba.bank_name,
    SUM(CASE WHEN tm.movement_type IN ('bank_deposit', 'check_received') THEN tm.amount ELSE -tm.amount END) AS net
FROM erp_treasury_movements tm
JOIN erp_bank_accounts ba ON ba.id = tm.bank_account_id
WHERE tm.tenant_id = $1 AND tm.status = 'confirmed' AND tm.bank_account_id IS NOT NULL
    AND tm.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM'), ba.bank_name
ORDER BY period;

-- name: AnalyticsPaymentMethods :many
SELECT
    COALESCE(payment_method, movement_type) AS method,
    COUNT(*) AS count,
    SUM(amount) AS total
FROM erp_treasury_movements
WHERE tenant_id = $1 AND status = 'confirmed'
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY COALESCE(payment_method, movement_type)
ORDER BY total DESC;

-- ─── Invoicing ─────────────────────────────────────────────────────────────

-- name: AnalyticsMonthlyTotals :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    direction,
    COUNT(*) AS count,
    SUM(total) AS total
FROM erp_invoices
WHERE tenant_id = $1 AND status IN ('posted', 'paid')
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM'), direction
ORDER BY period;

-- name: AnalyticsTaxSummary :many
SELECT
    period,
    direction,
    SUM(net_amount) AS net_total,
    SUM(tax_amount) AS tax_total
FROM erp_tax_entries
WHERE tenant_id = $1
    AND period BETWEEN sqlc.arg(period_from) AND sqlc.arg(period_to)
GROUP BY period, direction
ORDER BY period;

-- name: AnalyticsTopCustomers :many
SELECT
    e.id AS entity_id,
    e.name AS entity_name,
    COUNT(*) AS invoice_count,
    SUM(i.total) AS total
FROM erp_invoices i
JOIN erp_entities e ON e.id = i.entity_id
WHERE i.tenant_id = $1 AND i.direction = 'issued' AND i.status IN ('posted', 'paid')
    AND i.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY e.id, e.name
ORDER BY total DESC
LIMIT sqlc.arg(top_n);

-- name: AnalyticsTopSuppliers :many
SELECT
    e.id AS entity_id,
    e.name AS entity_name,
    COUNT(*) AS invoice_count,
    SUM(i.total) AS total
FROM erp_invoices i
JOIN erp_entities e ON e.id = i.entity_id
WHERE i.tenant_id = $1 AND i.direction = 'received' AND i.status IN ('posted', 'paid')
    AND i.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY e.id, e.name
ORDER BY total DESC
LIMIT sqlc.arg(top_n);

-- ─── Current Accounts ──────────────────────────────────────────────────────

-- name: AnalyticsAging :many
SELECT
    e.id AS entity_id, e.name AS entity_name,
    SUM(CASE WHEN am.balance > 0 AND i.due_date >= CURRENT_DATE THEN am.balance ELSE 0 END) AS current_amount,
    SUM(CASE WHEN am.balance > 0 AND i.due_date BETWEEN CURRENT_DATE - 30 AND CURRENT_DATE - 1 THEN am.balance ELSE 0 END) AS days_1_30,
    SUM(CASE WHEN am.balance > 0 AND i.due_date BETWEEN CURRENT_DATE - 60 AND CURRENT_DATE - 31 THEN am.balance ELSE 0 END) AS days_31_60,
    SUM(CASE WHEN am.balance > 0 AND i.due_date BETWEEN CURRENT_DATE - 90 AND CURRENT_DATE - 61 THEN am.balance ELSE 0 END) AS days_61_90,
    SUM(CASE WHEN am.balance > 0 AND i.due_date < CURRENT_DATE - 90 THEN am.balance ELSE 0 END) AS days_over_90
FROM erp_account_movements am
JOIN erp_entities e ON e.id = am.entity_id
LEFT JOIN erp_invoices i ON i.id = am.invoice_id
WHERE am.tenant_id = $1 AND am.direction = sqlc.arg(direction) AND am.balance > 0
GROUP BY e.id, e.name
ORDER BY SUM(am.balance) DESC;

-- name: AnalyticsCollectionRate :many
SELECT
    to_char(am.date, 'YYYY-MM') AS period,
    SUM(CASE WHEN am.movement_type = 'invoice' THEN am.amount ELSE 0 END) AS invoiced,
    SUM(CASE WHEN am.movement_type = 'payment' THEN am.amount ELSE 0 END) AS collected
FROM erp_account_movements am
WHERE am.tenant_id = $1 AND am.direction = 'receivable'
    AND am.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(am.date, 'YYYY-MM')
ORDER BY period;

-- ─── Stock ─────────────────────────────────────────────────────────────────

-- name: AnalyticsStockValuation :many
SELECT
    a.id, a.code, a.name,
    COALESCE(sl.quantity, 0) AS quantity,
    COALESCE(a.last_cost, 0) AS unit_cost,
    COALESCE(sl.quantity, 0) * COALESCE(a.last_cost, 0) AS stock_value
FROM erp_articles a
LEFT JOIN erp_stock_levels sl ON sl.article_id = a.id AND sl.tenant_id = a.tenant_id
WHERE a.tenant_id = $1 AND a.active = true AND COALESCE(sl.quantity, 0) > 0
ORDER BY stock_value DESC;

-- name: AnalyticsStockRotation :many
SELECT
    a.id, a.code, a.name,
    COALESCE(sl.quantity, 0) AS current_stock,
    COALESCE(consumed.qty, 0) AS consumed_qty,
    CASE WHEN (COALESCE(sl.quantity, 0) + COALESCE(consumed.qty, 0)) > 0
        THEN COALESCE(consumed.qty, 0) / ((COALESCE(sl.quantity, 0) + COALESCE(consumed.qty, 0)) / 2.0)
        ELSE 0
    END AS rotation_index
FROM erp_articles a
LEFT JOIN erp_stock_levels sl ON sl.article_id = a.id AND sl.tenant_id = a.tenant_id
LEFT JOIN (
    SELECT sm2.article_id, SUM(sm2.quantity) AS qty
    FROM erp_stock_movements sm2
    WHERE sm2.tenant_id = sqlc.arg(tenant_id) AND sm2.movement_type = 'out'
        AND sm2.created_at >= sqlc.arg(date_from)
    GROUP BY sm2.article_id
) consumed ON consumed.article_id = a.id
WHERE a.tenant_id = sqlc.arg(tenant_id) AND a.active = true
ORDER BY rotation_index DESC;

-- name: AnalyticsBelowMinimum :many
SELECT
    a.id, a.code, a.name,
    a.min_stock,
    COALESCE(sl.quantity, 0) AS current_stock,
    a.min_stock - COALESCE(sl.quantity, 0) AS deficit
FROM erp_articles a
LEFT JOIN erp_stock_levels sl ON sl.article_id = a.id AND sl.tenant_id = a.tenant_id
WHERE a.tenant_id = $1 AND a.active = true
    AND a.min_stock > 0 AND COALESCE(sl.quantity, 0) < a.min_stock
ORDER BY deficit DESC;

-- name: AnalyticsMovementSummary :many
SELECT
    sm.movement_type,
    COUNT(*) AS count,
    SUM(sm.quantity) AS total_quantity,
    SUM(sm.quantity * sm.unit_cost) AS total_value
FROM erp_stock_movements sm
WHERE sm.tenant_id = $1
    AND sm.created_at BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY sm.movement_type
ORDER BY total_value DESC;

-- ─── Production ────────────────────────────────────────────────────────────

-- name: AnalyticsProductionByStatus :many
SELECT status, COUNT(*) AS count
FROM erp_production_orders
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY status;

-- name: AnalyticsProductionByMonth :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    COUNT(*) AS total_orders,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed
FROM erp_production_orders
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsProductionEfficiency :many
SELECT
    COALESCE(pc.name, 'Sin centro') AS center_name,
    COUNT(*)::INT AS total_orders,
    COUNT(CASE WHEN po.status = 'completed' THEN 1 END)::INT AS completed
FROM erp_production_orders po
LEFT JOIN erp_production_centers pc ON pc.id = po.center_id
WHERE po.tenant_id = $1
    AND po.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY pc.name
ORDER BY total_orders DESC;

-- ─── Purchasing ────────────────────────────────────────────────────────────

-- name: AnalyticsPurchasingByMonth :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    COUNT(*) AS order_count,
    SUM(total) AS total
FROM erp_purchase_orders
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsSupplierRanking :many
SELECT
    e.id AS entity_id, e.name AS entity_name,
    COUNT(DISTINCT po.id) AS order_count,
    SUM(po.total) AS total_purchased,
    COALESCE(SUM(sd.points), 0) AS total_demerits
FROM erp_purchase_orders po
JOIN erp_entities e ON e.id = po.supplier_id
LEFT JOIN erp_supplier_demerits sd ON sd.supplier_id = e.id AND sd.tenant_id = e.tenant_id
WHERE po.tenant_id = $1
    AND po.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY e.id, e.name
ORDER BY total_purchased DESC
LIMIT sqlc.arg(top_n);

-- ─── Sales ─────────────────────────────────────────────────────────────────

-- name: AnalyticsSalesByMonth :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    COUNT(*) AS order_count,
    SUM(total) AS total
FROM erp_orders
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsQuotationConversion :many
SELECT
    to_char(q.date, 'YYYY-MM') AS period,
    COUNT(*) AS total_quotations,
    COUNT(*) FILTER (WHERE q.status = 'approved') AS converted
FROM erp_quotations q
WHERE q.tenant_id = $1
    AND q.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(q.date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsCustomerConcentration :many
SELECT
    e.id AS entity_id, e.name AS entity_name,
    SUM(o.total) AS total_sales
FROM erp_orders o
JOIN erp_entities e ON e.id = o.customer_id
WHERE o.tenant_id = $1 AND o.customer_id IS NOT NULL
    AND o.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY e.id, e.name
ORDER BY total_sales DESC
LIMIT sqlc.arg(top_n);

-- ─── HR ────────────────────────────────────────────────────────────────────

-- name: AnalyticsHeadcount :many
SELECT
    COALESCE(d.name, 'Sin departamento') AS department,
    COUNT(*) AS headcount
FROM erp_employee_details ed
JOIN erp_entities e ON e.id = ed.entity_id AND e.tenant_id = ed.tenant_id
LEFT JOIN erp_departments d ON d.id = ed.department_id
WHERE ed.tenant_id = $1 AND e.active = true
    AND (ed.termination_date IS NULL OR ed.termination_date > CURRENT_DATE)
GROUP BY d.name
ORDER BY headcount DESC;

-- name: AnalyticsAbsencesByMonth :many
SELECT
    to_char(date_from, 'YYYY-MM') AS period,
    event_type,
    COUNT(*) AS count
FROM erp_hr_events
WHERE tenant_id = $1
    AND event_type IN ('absence', 'leave', 'vacation', 'accident')
    AND date_from BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date_from, 'YYYY-MM'), event_type
ORDER BY period;

-- name: AnalyticsOvertimeSummary :many
SELECT
    to_char(date_from, 'YYYY-MM') AS period,
    COUNT(*) AS event_count,
    SUM(COALESCE(hours, 0)) AS total_hours
FROM erp_hr_events
WHERE tenant_id = $1 AND event_type = 'overtime'
    AND date_from BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date_from, 'YYYY-MM')
ORDER BY period;

-- ─── Quality ───────────────────────────────────────────────────────────────

-- name: AnalyticsNCByType :many
SELECT severity, status, COUNT(*) AS count
FROM erp_nonconformities
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY severity, status
ORDER BY count DESC;

-- name: AnalyticsNCResolutionTime :many
SELECT
    severity,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'closed') AS closed,
    AVG(EXTRACT(EPOCH FROM (closed_at - created_at)) / 86400)
        FILTER (WHERE status = 'closed' AND closed_at IS NOT NULL)
        AS avg_days_to_close
FROM erp_nonconformities
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY severity;

-- ─── Maintenance ───────────────────────────────────────────────────────────

-- name: AnalyticsWOCompletionRate :many
SELECT
    to_char(date, 'YYYY-MM') AS period,
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed
FROM erp_work_orders
WHERE tenant_id = $1
    AND date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
GROUP BY to_char(date, 'YYYY-MM')
ORDER BY period;

-- name: AnalyticsCostByAsset :many
SELECT
    ma.id AS asset_id, ma.code AS asset_code, ma.name AS asset_name,
    COUNT(wo.id) AS wo_count,
    COALESCE(SUM(p.quantity * a.last_cost), 0) AS total_cost
FROM erp_maintenance_assets ma
LEFT JOIN erp_work_orders wo ON wo.asset_id = ma.id AND wo.tenant_id = ma.tenant_id
    AND wo.date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to)
LEFT JOIN erp_work_order_parts p ON p.work_order_id = wo.id
LEFT JOIN erp_articles a ON a.id = p.article_id
WHERE ma.tenant_id = $1 AND ma.active = true
GROUP BY ma.id, ma.code, ma.name
ORDER BY total_cost DESC;

-- ─── Dashboard KPIs ────────────────────────────────────────────────────────

-- name: KPIMonthRevenue :one
SELECT COALESCE(SUM(total), 0) AS revenue
FROM erp_invoices
WHERE tenant_id = $1 AND direction = 'issued' AND status IN ('posted', 'paid')
    AND date >= date_trunc('month', CURRENT_DATE);

-- name: KPIMonthExpense :one
SELECT COALESCE(SUM(total), 0) AS expense
FROM erp_invoices
WHERE tenant_id = $1 AND direction = 'received' AND status IN ('posted', 'paid')
    AND date >= date_trunc('month', CURRENT_DATE);

-- name: KPICashBalance :one
SELECT COALESCE(SUM(
    CASE WHEN movement_type IN ('cash_in', 'bank_deposit', 'check_received') THEN amount ELSE -amount END
), 0) AS balance
FROM erp_treasury_movements
WHERE tenant_id = $1 AND status = 'confirmed';

-- name: KPIAccountsReceivable :one
SELECT COALESCE(SUM(balance), 0) AS total
FROM erp_account_movements
WHERE tenant_id = $1 AND direction = 'receivable' AND balance > 0;

-- name: KPIAccountsPayable :one
SELECT COALESCE(SUM(balance), 0) AS total
FROM erp_account_movements
WHERE tenant_id = $1 AND direction = 'payable' AND balance > 0;

-- name: KPIActiveProdOrders :one
SELECT COUNT(*) AS count
FROM erp_production_orders
WHERE tenant_id = $1 AND status IN ('planned', 'in_progress');

-- name: KPIPendingPurchases :one
SELECT COUNT(*) AS count
FROM erp_purchase_orders
WHERE tenant_id = $1 AND status IN ('draft', 'approved', 'partial');

-- name: KPIOpenQuotations :one
SELECT COUNT(*) AS count
FROM erp_quotations
WHERE tenant_id = $1 AND status IN ('draft', 'sent');

-- name: KPIStockBelowMin :one
SELECT COUNT(*) AS count
FROM erp_articles a
LEFT JOIN erp_stock_levels sl ON sl.article_id = a.id AND sl.tenant_id = a.tenant_id
WHERE a.tenant_id = $1 AND a.active = true
    AND a.min_stock > 0 AND COALESCE(sl.quantity, 0) < a.min_stock;

-- name: KPIHeadcount :one
SELECT COUNT(*) AS count
FROM erp_employee_details ed
JOIN erp_entities e ON e.id = ed.entity_id AND e.tenant_id = ed.tenant_id
WHERE ed.tenant_id = $1 AND e.active = true
    AND (ed.termination_date IS NULL OR ed.termination_date > CURRENT_DATE);

-- name: KPIAbsencesThisMonth :one
SELECT COUNT(*) AS count
FROM erp_hr_events
WHERE tenant_id = $1 AND event_type IN ('absence', 'leave')
    AND date_from >= date_trunc('month', CURRENT_DATE);

-- name: KPIOpenNonconformities :one
SELECT COUNT(*) AS count
FROM erp_nonconformities
WHERE tenant_id = $1 AND status != 'closed';

-- name: KPIPendingWorkOrders :one
SELECT COUNT(*) AS count
FROM erp_work_orders
WHERE tenant_id = $1 AND status IN ('open', 'in_progress');

-- ─── Pending Exports ───────────────────────────────────────────────────────

-- name: CreatePendingExport :one
INSERT INTO erp_pending_exports (tenant_id, user_id, export_name, row_count, format, params, columns_def)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPendingExport :one
SELECT * FROM erp_pending_exports WHERE id = $1 AND tenant_id = $2;

-- name: ListUserExports :many
SELECT * FROM erp_pending_exports
WHERE tenant_id = $1 AND user_id = $2
ORDER BY requested_at DESC
LIMIT 20;

-- name: MarkExportRunning :exec
UPDATE erp_pending_exports SET status = 'running', started_at = now() WHERE id = $1;

-- name: MarkExportReady :exec
UPDATE erp_pending_exports SET status = 'ready', file_key = $2, ready_at = now() WHERE id = $1;

-- name: MarkExportFailed :exec
UPDATE erp_pending_exports SET status = 'failed', error = $2 WHERE id = $1;

-- name: ListStaleRunningExports :many
SELECT * FROM erp_pending_exports
WHERE status = 'running' AND started_at < now() - sqlc.arg(stale_threshold)::INTERVAL;

-- name: DeleteExpiredExports :exec
DELETE FROM erp_pending_exports WHERE expires_at < now();
