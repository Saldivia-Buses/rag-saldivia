// Accounting
export interface Account {
  id: string;
  code: string;
  name: string;
  account_type: string;
  is_detail: boolean;
  active: boolean;
  parent_id: string | null;
}

export interface JournalEntry {
  id: string;
  number: string;
  date: string;
  concept: string;
  entry_type: string;
  status: string;
  user_id: string;
  created_at: string;
}

export interface JournalLine {
  account_code: string;
  account_name: string;
  debit: number;
  credit: number;
  description: string;
}

export interface AccountBalance {
  account_code: string;
  account_name: string;
  total_debit: number;
  total_credit: number;
  balance: number;
}

export interface FiscalYear {
  id: string;
  year: number;
  start_date: string;
  end_date: string;
  status: string;
  closed_at: string | null;
}

export interface CostCenter {
  id: string;
  code: string;
  name: string;
  active: boolean;
}

// Treasury
export interface TreasuryMovement {
  id: string;
  date: string;
  number: string;
  movement_type: string;
  amount: number;
  entity_name: string | null;
  notes: string;
  status: string;
}

export interface Check {
  id: string;
  direction: string;
  number: string;
  bank_name: string;
  amount: number;
  issue_date: string;
  due_date: string;
  status: string;
}

export interface BankBalance {
  bank_name: string;
  account_number: string;
  total_in: number;
  total_out: number;
  balance: number;
}

// Quality audits (erp_audits — Phase 1 §UI)
export interface QualityAudit {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  audit_type: string;
  scope: string;
  lead_auditor_id: string | null;
  status: string;
  score: number | null;
  notes: string;
  created_at: string;
}

// Controlled documents (erp_controlled_documents — Phase 1 §UI)
export interface ControlledDocument {
  id: string;
  tenant_id: string;
  code: string;
  title: string;
  revision: number;
  doc_type_id: string | null;
  file_key: string;
  approved_by: string | null;
  approved_at: string | null;
  status: string;
  created_at: string;
}

// Quality action plans (erp_quality_action_plans — Phase 1 §UI)
export interface ActionPlan {
  id: string;
  tenant_id: string;
  nonconformity_id: string | null;
  responsible_id: string | null;
  section_id: string | null;
  description: string;
  planned_start: string | null;
  target_date: string | null;
  closed_date: string | null;
  time_savings_hours: number | null;
  cost_savings: number | null;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// Quality indicators (erp_quality_indicators — Phase 1 §UI)
export interface QualityIndicator {
  id: string;
  tenant_id: string;
  period: string;
  indicator_type: string;
  value: number | null;
  target: number | null;
  created_at: string;
}

// Stock movement row (erp_stock_movements — Phase 1 §UI)
export interface StockMovement {
  id: string;
  tenant_id: string;
  article_id: string;
  article_code: string | null;
  article_name: string | null;
  warehouse_id: string | null;
  warehouse_name: string | null;
  movement_type: string;
  quantity: number;
  unit_cost: number | null;
  reference_type: string;
  reference_id: string | null;
  movement_date: string | null;
  user_id: string;
  notes: string;
  created_at: string;
}

// Article cost row (erp_article_costs — Phase 1 §UI)
export interface ArticleCost {
  id: string;
  tenant_id: string;
  article_id: string;
  cost: number;
  percentage_1: number | null;
  percentage_2: number | null;
  percentage_3: number | null;
  supplier_article_code: string;
  supplier_code: string;
  supplier_entity_id: string | null;
  invoice_date: string | null;
  last_update_date: string | null;
  movement_no: number;
  movement_post: number;
  movement_date: string | null;
  recalc_flag: number;
  created_at: string;
}

// Invoice notes (REG_MOVIMIENTO_OBS parity — Phase 1 §UI)
export interface InvoiceNote {
  id: string;
  tenant_id: string;
  legacy_id: number;
  observation_date: string | null;
  observation_time: string | null;
  observation: string;
  invoice_id: string | null;
  invoice_legacy_id: number;
  login: string;
  contact_legacy_id: number;
  source_table: string;
  system_code: string;
  movement_date: string | null;
  account_code: number;
  concept_code: number;
  movement_voucher_class: number;
  movement_no: number;
  created_at: string;
}

// Entity credit ratings (REG_CUENTA_CALIFICACION parity — Phase 1 §UI)
export interface EntityCreditRating {
  id: string;
  tenant_id: string;
  legacy_id: number;
  entity_id: string | null;
  entity_legacy_id: number;
  rating: string; // A | B | C | X | '-'
  rated_at: string | null;
  reference: string;
  created_at: string;
  entity_name: string | null;
  entity_type: string | null;
}

// Archived cheque history (CARCHEHI parity — Phase 1 §UI)
export interface CheckHistoryEntry {
  id: string;
  tenant_id: string;
  legacy_id: number;
  legacy_carint: number;
  check_type: number;
  number: string;
  bank_name: string;
  amount: number;
  operation_date: string | null;
  credited_at: string | null;
  returned_at: string | null;
  issue_date: string | null;
  description: string;
  observation: string;
  reference: string;
  owner_ident: string;
  accredited: number;
  entity_legacy_code: number;
  entity_id: string | null;
  movement_no: number;
  pay_no: number;
  received_no: number;
  branch: number;
  created_at: string;
}

// Bank-import staging rows (BCS_IMPORTACION parity — Phase 1 §UI)
export interface BankImport {
  id: string;
  tenant_id: string;
  legacy_id: number;
  movement_date: string | null;
  concept_name: string;
  movement_no: number;
  amount: number;
  debit: number;
  credit: number;
  balance: number;
  movement_code: string;
  treasury_movement_id: string | null;
  treasury_legacy_id: number;
  imported_at: string | null;
  account_number: number;
  account_entity_id: string | null;
  processed: number; // 0 = pendiente, 1 = procesado, 2 = anulado
  comments: string;
  internal_no: number;
  branch: string;
  created_at: string;
}

// Invoicing
export interface Invoice {
  id: string;
  number: string;
  date: string;
  invoice_type: string;
  direction: string;
  entity_name: string;
  total: number;
  status: string;
}

export interface Withholding {
  id: string;
  entity_name: string;
  type: string;
  rate: number;
  base_amount: number;
  amount: number;
  date: string;
}

export interface Reconciliation {
  id: string;
  bank_account_id: string;
  bank_name: string;
  period: string;
  statement_balance: number;
  book_balance: number;
  status: string;
}

export interface Receipt {
  id: string;
  number: string;
  date: string;
  receipt_type: string;
  entity_name: string;
  total: number;
  status: string;
}

export interface TaxBookEntry {
  invoice_number: string;
  date: string;
  entity_name: string;
  net_amount: number;
  tax_amount: number;
  total: number;
  direction: string;
}

export interface QCInspection {
  id: string;
  receipt_number: string;
  article_name: string;
  quantity: number;
  accepted_qty: number;
  rejected_qty: number;
  status: string;
}

// Current Accounts
export interface EntityBalance {
  entity_id: string;
  entity_name: string;
  entity_type: string;
  direction: string;
  open_balance: number;
}

export interface OverdueInvoice {
  entity_name: string;
  invoice_number: string;
  due_date: string;
  amount: number;
  balance: number;
}

// Entity search (picker) — subset of ListEntitiesRow used by UI pickers.
export interface EntitySearchResult {
  id: string;
  type: string;
  code: string;
  name: string;
  email: string | null;
  phone: string | null;
  active: boolean;
}

// Payment Complaints (RECLAMOPAGOS parity — Phase 1 §UI)
export interface PaymentComplaint {
  id: string;
  tenant_id: string;
  legacy_id: number;
  complaint_date: string | null;
  entity_legacy_code: number;
  entity_id: string | null;
  observation: string;
  status_flag: number; // 0 = pendiente, 1 = cumplida
  login: string;
  created_at: string;
}
