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

// ─── 2.0.17 §UI parity clusters ─────────────────────────────────────────

export interface ChassisBrand {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  active: boolean;
}

export interface ChassisModel {
  id: string;
  tenant_id: string;
  brand_id: string | null;
  model_code: string;
  description: string;
  traction: string;
  engine_location: string;
  active: boolean;
}

export interface RiskAgent {
  id: string;
  tenant_id: string;
  name: string;
  risk_type: string;
  active: boolean;
  created_at: string;
}

export interface RiskExposure {
  id: string;
  tenant_id: string;
  entity_id: string | null;
  agent_id: string | null;
  level: string;
  hours_per_day: number | null;
  mitigation: string;
  created_at: string;
}

export interface Attendance {
  id: string;
  tenant_id: string;
  entity_id: string | null;
  date: string | null;
  clock_in: string | null;
  clock_out: string | null;
  hours: number | null;
  source: string;
}

export interface Communication {
  id: string;
  tenant_id: string;
  subject: string;
  body: string;
  sender_id: string;
  priority: string;
  created_at: string;
}

export interface CalendarEvent {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  start_at: string | null;
  end_at: string | null;
  all_day: boolean;
  entity_id: string | null;
}

export interface Survey {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  status: string;
  user_id: string;
  created_at: string;
}

export interface ProductSection {
  id: string;
  tenant_id: string;
  legacy_id: number;
  name: string;
  sort_order: number;
  rubro_id: number;
  active: boolean;
  created_at: string;
}

export interface Product {
  id: string;
  tenant_id: string;
  legacy_id: number;
  description: string;
  supplier_entity_id: string | null;
  supplier_code: number;
  created_at: string;
}

export interface ProductAttribute {
  id: string;
  tenant_id: string;
  legacy_id: number;
  name: string;
  attribute_type: string;
  section_id: string | null;
  section_legacy_id: number;
  article_code: string;
  active: boolean;
  sort_order: number;
}

export interface Tool {
  id: string;
  tenant_id: string;
  legacy_id: number;
  code: string;
  article_code: string;
  article_id: string | null;
  inventory_code: string;
  name: string;
  status: string;
  observation: string;
  purchase_order_date: string | null;
  delivery_date: string | null;
  supplier_code: string;
  supplier_entity_id: string | null;
  created_at: string;
}

export interface MaintenanceAsset {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  asset_type: string;
  unit_id: string | null;
  location: string;
  active: boolean;
  created_at: string;
}

// ─── 2.0.16 §UI parity clusters ─────────────────────────────────────────

export interface BankReconciliation {
  id: string;
  tenant_id: string;
  bank_account_id: string;
  period: string;
  statement_balance: number | null;
  book_balance: number | null;
  status: string;
  user_id?: string;
  confirmed_at: string | null;
  created_at: string;
  bank_name: string;
  account_number: string;
}

export interface BankStatementLine {
  id: string;
  tenant_id: string;
  reconciliation_id: string;
  date: string | null;
  description: string;
  amount: number | null;
  reference: string;
  matched: boolean;
  movement_id: string | null;
  created_at: string;
}

export interface ReconciliationDetail {
  reconciliation: BankReconciliation;
  lines: BankStatementLine[];
}

export interface PurchaseOrder {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  supplier_id: string | null;
  status: string;
  currency_id: string | null;
  total: number | null;
  notes: string;
  user_id: string;
  created_at: string;
  supplier_name?: string;
}

export interface PurchaseOrderLine {
  id: string;
  tenant_id: string;
  order_id: string;
  article_id: string | null;
  quantity: number | null;
  unit_price: number | null;
  received_qty: number | null;
  sort_order: number;
  article_code: string;
  article_name: string;
}

export interface PurchaseOrderDetail {
  order: PurchaseOrder;
  lines: PurchaseOrderLine[];
}

export interface Warehouse {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  location: string;
  active: boolean;
}

export interface CashCount {
  id: string;
  tenant_id: string;
  cash_register_id: string;
  date: string | null;
  expected: number | null;
  counted: number | null;
  difference: number | null;
  user_id: string;
  notes: string;
  created_at: string;
}

export interface CustomerVehicle {
  id: string;
  tenant_id: string;
  owner_id: string | null;
  driver_id: string | null;
  manufacturing_unit_id: string | null;
  plate: string;
  chassis_serial: string;
  body_serial: string;
  internal_number: number | null;
  brand: string;
  model_year: number | null;
  seating_capacity: number;
  fuel_type: string;
  color: string;
  purchase_date: string | null;
  purchase_price: number | null;
  warranty_months: number;
  destination: string;
  observations: string;
  active: boolean;
}

export interface VehicleIncident {
  id: string;
  tenant_id: string;
  vehicle_id: string | null;
  incident_type_id: string | null;
  incident_date: string | null;
  location: string;
  responsible: string;
  notes: string;
  status: string;
  resolved_at: string | null;
  created_at: string;
  updated_at: string;
  incident_type_name: string | null;
}

export interface ProductionCenter {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  active: boolean;
}

export interface BankAccount {
  id: string;
  tenant_id: string;
  bank_name: string;
  branch: string;
  account_number: string;
  cbu: string | null;
  alias: string | null;
  currency_id: string | null;
  account_id: string | null;
  active: boolean;
  created_at: string;
}

export interface CashRegister {
  id: string;
  tenant_id: string;
  name: string;
  account_id: string | null;
  active: boolean;
  created_at: string;
}

export interface PriceList {
  id: string;
  tenant_id: string;
  name: string;
  currency_id: string | null;
  valid_from: string | null;
  valid_until: string | null;
  active: boolean;
}

export interface PriceListItem {
  id: string;
  tenant_id: string;
  price_list_id: string;
  article_id: string | null;
  description: string | null;
  price: number | null;
  article_code: string | null;
  article_name: string | null;
}

export interface PriceListDetail {
  price_list: PriceList;
  items: PriceListItem[];
}

export interface Quotation {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  customer_id: string | null;
  status: string;
  currency_id: string | null;
  total: number | null;
  valid_until: string | null;
  notes: string;
  user_id: string;
  created_at: string;
  customer_name?: string;
}

export interface QuotationLine {
  id: string;
  tenant_id: string;
  quotation_id: string;
  article_id: string | null;
  description: string;
  quantity: number | null;
  unit_price: number | null;
  sort_order: number;
  metadata?: string | null;
}

export interface QuotationOption {
  id: string;
  tenant_id: string;
  legacy_id: number;
  quotation_id: string;
  quotation_legacy_id: number;
  section_legacy_id: number;
  description: string;
  created_at: string;
}

export interface QuotationDetail {
  quotation: Quotation;
  lines: QuotationLine[];
  options: QuotationOption[];
}

export interface MaintenanceAsset {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  asset_type: string;
  unit_id: string | null;
  location: string;
  active: boolean;
  created_at: string;
}

export interface MaintenancePlan {
  id: string;
  tenant_id: string;
  asset_id: string;
  name: string;
  frequency_days: number | null;
  frequency_km: number | null;
  frequency_hours: number | null;
  last_done: string | null;
  next_due: string | null;
  active: boolean;
}

export interface WorkOrder {
  id: string;
  tenant_id: string;
  number: string;
  asset_id: string | null;
  plan_id?: string | null;
  date: string | null;
  work_type: string;
  description: string;
  assigned_to: string | null;
  status: string;
  priority: string;
  completed_at: string | null;
  user_id: string;
  notes: string;
  created_at: string;
  asset_code?: string;
  asset_name?: string;
}

export interface WorkOrderPart {
  id: string;
  tenant_id: string;
  work_order_id: string;
  article_id: string;
  quantity: number | null;
  article_code: string;
  article_name: string;
}

export interface WorkOrderDetail {
  order: WorkOrder;
  parts: WorkOrderPart[];
}

export interface QCInspection {
  id: string;
  tenant_id: string;
  receipt_id: string | null;
  receipt_line_id?: string | null;
  article_id: string | null;
  quantity: number | null;
  accepted_qty: number | null;
  rejected_qty: number | null;
  status: string;
  inspector_id: string;
  notes?: string;
  completed_at?: string | null;
  created_at: string;
  article_code?: string;
  article_name?: string;
  receipt_number?: string;
}

export interface Entity {
  id: string;
  tenant_id: string;
  type: string;
  code: string;
  name: string;
  tax_id_hash?: string | null;
  email?: string | null;
  phone?: string | null;
  active: boolean;
  created_at: string;
  updated_at?: string;
}

export interface EntityContact {
  id: string;
  tenant_id: string;
  entity_id: string;
  name: string;
  role?: string;
  email?: string | null;
  phone?: string | null;
  primary: boolean;
}

export interface EntityNote {
  id: string;
  tenant_id: string;
  entity_id: string;
  user_id: string;
  note: string;
  created_at: string;
}

export interface EntityDocument {
  id: string;
  tenant_id: string;
  entity_id: string;
  doc_type: string;
  filename: string;
  created_at: string;
}

export interface EntityDetail {
  entity: Entity;
  contacts: EntityContact[];
  documents: EntityDocument[];
  notes: EntityNote[];
  relations: unknown[];
}

export interface SupplierDemerit {
  id: string;
  tenant_id: string;
  supplier_id: string;
  inspection_id: string | null;
  points: number;
  reason: string;
  created_at: string;
}

export interface SalesOrder {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  order_type: string;
  customer_id: string | null;
  quotation_id: string | null;
  status: string;
  total: number | null;
  user_id: string;
  notes: string;
  created_at: string;
  customer_name?: string | null;
}

export interface ProductionOrder {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  product_id: string | null;
  center_id: string | null;
  quantity: number | null;
  status: string;
  priority: number;
  order_id: string | null;
  start_date: string | null;
  end_date: string | null;
  user_id: string;
  notes: string;
  created_at: string;
  product_code?: string;
  product_name?: string;
}

export interface ProductionMaterial {
  id: string;
  tenant_id: string;
  order_id: string;
  article_id: string;
  required_qty: number | null;
  consumed_qty: number | null;
  warehouse_id: string | null;
  article_code: string;
  article_name: string;
}

export interface ProductionStep {
  id: string;
  tenant_id: string;
  order_id: string;
  step_name: string;
  sort_order: number;
  status: string;
  assigned_to: string | null;
  started_at: string | null;
  completed_at: string | null;
  notes: string;
}

export interface ProductionInspection {
  id: string;
  tenant_id: string;
  order_id: string;
  step_id: string | null;
  inspector_id: string;
  result: string;
  observations: string;
  created_at: string;
}

export interface ProductionOrderDetail {
  order: ProductionOrder;
  materials: ProductionMaterial[];
  steps: ProductionStep[];
  inspections: ProductionInspection[];
}

export interface FuelLog {
  id: string;
  tenant_id: string;
  asset_id: string | null;
  date: string | null;
  liters: number | null;
  km_reading: number | null;
  cost: number | null;
  user_id: string;
  created_at: string;
  asset_code?: string;
  asset_name?: string;
}

export interface StockArticle {
  id: string;
  code: string;
  name: string;
  article_type: string;
  min_stock: number;
  avg_cost: number;
  active: boolean;
}

export interface BOMEntry {
  id: string;
  tenant_id: string;
  parent_id: string;
  child_id: string;
  quantity: number | null;
  unit_id: string | null;
  sort_order: number;
  notes: string;
  child_code: string;
  child_name: string;
}

export interface SupplierScorecard {
  id: string;
  tenant_id: string;
  supplier_id: string;
  period: string;
  total_receipts: number;
  accepted_qty: number | null;
  rejected_qty: number | null;
  total_demerits: number;
  quality_score: number | null;
  created_at: string;
  supplier_name: string;
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
  legacy_id: number;
  article_code: string;
  article_id: string;
  subsystem_code: string;
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
  article_name: string | null;
  supplier_name: string | null;
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
  entity_name?: string;
  entity_id?: string;
  currency_id?: string | null;
  subtotal?: number | null;
  tax_amount?: number | null;
  total: number;
  due_date?: string | null;
  order_id?: string | null;
  journal_entry_id?: string | null;
  afip_cae?: string | null;
  afip_cae_due?: string | null;
  user_id?: string;
  created_at?: string;
  status: string;
}

export interface InvoiceLine {
  id: string;
  tenant_id: string;
  invoice_id: string;
  article_id: string | null;
  description: string;
  quantity: number | null;
  unit_price: number | null;
  tax_rate: number | null;
  tax_amount: number | null;
  line_total: number | null;
  sort_order: number;
}

export interface InvoiceDetail {
  invoice: Invoice;
  lines: InvoiceLine[];
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
  entity_id?: string;
  entity_name: string;
  total: number;
  journal_entry_id?: string | null;
  user_id?: string;
  notes?: string;
  status: string;
  created_at?: string;
}

export interface ReceiptPayment {
  id: string;
  tenant_id: string;
  receipt_id: string;
  payment_method: string;
  amount: number | null;
  treasury_movement_id: string | null;
  check_id: string | null;
  bank_account_id: string | null;
  notes: string;
}

export interface ReceiptAllocation {
  id: string;
  tenant_id: string;
  receipt_id: string;
  invoice_id: string;
  amount: number | null;
  invoice_number: string;
  invoice_total: number | null;
}

export interface ReceiptDetail {
  receipt: Receipt;
  payments: ReceiptPayment[];
  allocations: ReceiptAllocation[];
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

// Employee list row (erp_employee_details + erp_entities — Phase 1 §UI)
export interface EmployeeListRow {
  id: string;
  tenant_id: string;
  entity_id: string;
  department_id: string | null;
  position: string;
  hire_date: string | null;
  termination_date: string | null;
  schedule_type: string;
  created_at: string;
  entity_name: string;
  entity_code: string;
}

// Employee detail (erp_employee_details — Phase 1 §UI).
// Note: GET /v1/erp/hr/employees/{id} takes entity_id as the URL param.
export interface EmployeeDetail {
  id: string;
  tenant_id: string;
  entity_id: string;
  department_id: string | null;
  position: string;
  hire_date: string | null;
  termination_date: string | null;
  union_id: string | null;
  health_plan_id: string | null;
  schedule_type: string;
  category_id: string | null;
  metadata: unknown;
  created_at: string;
  updated_at: string;
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
