export const erpKeys = {
  all: ["erp"] as const,
  catalogs: (type?: string) => [...erpKeys.all, "catalogs", { type }] as const,
  entities: (type?: string, search?: string) =>
    [...erpKeys.all, "entities", { type, search }] as const,
  creditRatings: (params?: Record<string, string>) =>
    [...erpKeys.all, "entities", "credit-ratings", params] as const,
  accounts: () => [...erpKeys.all, "accounts"] as const,
  entries: (params?: Record<string, string>) =>
    [...erpKeys.all, "entries", params] as const,
  entry: (id: string) => [...erpKeys.all, "entries", id] as const,
  balance: () => [...erpKeys.all, "balance"] as const,
  ledger: (accountId?: string) =>
    [...erpKeys.all, "ledger", { accountId }] as const,
  fiscalYears: () => [...erpKeys.all, "fiscal-years"] as const,
  stockArticles: (params?: Record<string, string>) =>
    [...erpKeys.all, "stock", "articles", params] as const,
  stockLevels: () => [...erpKeys.all, "stock", "levels"] as const,
  warehouses: () => [...erpKeys.all, "stock", "warehouses"] as const,
  purchaseOrders: (status?: string) =>
    [...erpKeys.all, "purchasing", "orders", { status }] as const,
  quotations: (status?: string) =>
    [...erpKeys.all, "sales", "quotations", { status }] as const,
  invoices: (params?: Record<string, string>) =>
    [...erpKeys.all, "invoicing", "invoices", params] as const,
  invoice: (id: string) =>
    [...erpKeys.all, "invoicing", "invoices", id] as const,
  withholdings: () => [...erpKeys.all, "invoicing", "withholdings"] as const,
  invoiceNotes: (params?: Record<string, string>) =>
    [...erpKeys.all, "invoicing", "invoice-notes", params] as const,
  qualityAudits: (params?: Record<string, string>) =>
    [...erpKeys.all, "quality", "audits", params] as const,
  controlledDocuments: (params?: Record<string, string>) =>
    [...erpKeys.all, "quality", "documents", params] as const,
  actionPlans: (params?: Record<string, string>) =>
    [...erpKeys.all, "quality", "action-plans", params] as const,
  qualityIndicators: (params?: Record<string, string>) =>
    [...erpKeys.all, "quality", "indicators", params] as const,
  stockMovements: (params?: Record<string, string>) =>
    [...erpKeys.all, "stock", "movements", params] as const,
  articleCosts: (params?: Record<string, string>) =>
    [...erpKeys.all, "stock", "article-costs", params] as const,
  bankReconciliations: () =>
    [...erpKeys.all, "treasury", "reconciliations"] as const,
  warehouses2: () => [...erpKeys.all, "stock", "warehouses", "v2"] as const,
  cashCounts: (params?: Record<string, string>) =>
    [...erpKeys.all, "treasury", "cash-counts", params] as const,
  customerVehicles: (params?: Record<string, string>) =>
    [...erpKeys.all, "workshop", "vehicles", params] as const,
  vehicleIncidents: (params?: Record<string, string>) =>
    [...erpKeys.all, "workshop", "incidents", params] as const,
  productionCenters: () =>
    [...erpKeys.all, "production", "centers"] as const,
  bankAccountsCatalog: () =>
    [...erpKeys.all, "treasury", "bank-accounts", "catalog"] as const,
  cashRegisters: () =>
    [...erpKeys.all, "treasury", "cash-registers"] as const,
  priceLists: () => [...erpKeys.all, "sales", "price-lists"] as const,
  supplierScorecards: (params?: Record<string, string>) =>
    [...erpKeys.all, "quality", "scorecards", params] as const,
  // 2.0.17
  chassisBrands: () => [...erpKeys.all, "manufacturing", "chassis-brands"] as const,
  chassisModels: () => [...erpKeys.all, "manufacturing", "chassis-models"] as const,
  riskAgents: () => [...erpKeys.all, "safety", "risk-agents"] as const,
  riskExposures: () => [...erpKeys.all, "safety", "risk-exposures"] as const,
  attendance: (params?: Record<string, string>) =>
    [...erpKeys.all, "hr", "attendance", params] as const,
  communications: () => [...erpKeys.all, "admin", "communications"] as const,
  calendar: (params?: Record<string, string>) =>
    [...erpKeys.all, "admin", "calendar", params] as const,
  surveys: () => [...erpKeys.all, "admin", "surveys"] as const,
  productSections: () => [...erpKeys.all, "products", "sections"] as const,
  products: (params?: Record<string, string>) =>
    [...erpKeys.all, "products", "list", params] as const,
  productAttributes: (activeOnly?: boolean) =>
    [...erpKeys.all, "products", "attributes", { activeOnly }] as const,
  tools: (params?: Record<string, string>) =>
    [...erpKeys.all, "tools", "list", params] as const,
  maintenanceAssets: () =>
    [...erpKeys.all, "maintenance", "assets"] as const,
  treasuryMovements: () =>
    [...erpKeys.all, "treasury", "movements"] as const,
  treasuryBalance: () =>
    [...erpKeys.all, "treasury", "balance"] as const,
  bankAccounts: () => [...erpKeys.all, "treasury", "bank-accounts"] as const,
  checks: () => [...erpKeys.all, "treasury", "checks"] as const,
  bankImports: (params?: Record<string, string>) =>
    [...erpKeys.all, "treasury", "imports", params] as const,
  checkHistory: (params?: Record<string, string>) =>
    [...erpKeys.all, "treasury", "check-history", params] as const,
  receipts: (type?: string) =>
    [...erpKeys.all, "treasury", "receipts", { type }] as const,
  accountBalances: () =>
    [...erpKeys.all, "accounts", "balances"] as const,
  accountOverdue: () =>
    [...erpKeys.all, "accounts", "overdue"] as const,
  accountStatement: (entityId?: string) =>
    [...erpKeys.all, "accounts", "statement", { entityId }] as const,
  paymentComplaints: (params?: Record<string, string>) =>
    [...erpKeys.all, "accounts", "complaints", params] as const,
  productionOrders: (status?: string) =>
    [...erpKeys.all, "production", "orders", { status }] as const,
  employees: () => [...erpKeys.all, "hr", "employees"] as const,
  workOrders: (status?: string) =>
    [...erpKeys.all, "maintenance", "work-orders", { status }] as const,
  analytics: (domain: string, report: string, params?: Record<string, string>) =>
    [...erpKeys.all, "analytics", domain, report, params] as const,
  dashboardKPIs: () => [...erpKeys.all, "analytics", "dashboard", "kpis"] as const,
};
