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
