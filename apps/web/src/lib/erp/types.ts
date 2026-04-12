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
