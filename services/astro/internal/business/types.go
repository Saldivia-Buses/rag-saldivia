package business

// TimingWindow is a scored date range for business activity.
type TimingWindow struct {
	Counterparty string   `json:"counterparty"`
	Month        int      `json:"month"`
	DayStart     int      `json:"day_start"`
	DayEnd       int      `json:"day_end"`
	Score        float64  `json:"score"`   // 0-100
	Nature       string   `json:"nature"`  // "favorable", "caution", "avoid"
	Factors      []string `json:"factors"`
}

// CashFlowMonth holds monthly cash flow forecast.
type CashFlowMonth struct {
	Month    int     `json:"month"`
	Label    string  `json:"label"`    // "Enero", "Febrero", ...
	Inflow   float64 `json:"inflow"`   // 0-100 relative score
	Outflow  float64 `json:"outflow"`  // 0-100 relative score
	Net      float64 `json:"net"`      // inflow - outflow
	Rating   string  `json:"rating"`   // "abundancia", "neutro", "presión"
	Details  string  `json:"details"`  // brief astrological basis
}

// RiskCell is one cell in the risk heatmap (category × month).
type RiskCell struct {
	Category string `json:"category"` // "financiero", "operativo", "legal", "personal", "reputacional"
	Month    int    `json:"month"`
	Level    int    `json:"level"`    // 0-5 (0=none, 5=critical)
	Alert    string `json:"alert,omitempty"`
}

// AgendaItem is a single item in the daily/monthly agenda.
type AgendaItem struct {
	Date        string  `json:"date"` // YYYY-MM-DD
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`    // relevance score
	Category    string  `json:"category"` // "negotiation", "meeting", "deadline", "alert"
	Source      string  `json:"source"`   // "transit", "timing", "follow-up"
	Pinned      bool    `json:"pinned"`   // from user follow-ups
}

// QuarterlyForecast holds a 3-month outlook.
type QuarterlyForecast struct {
	Quarter    int      `json:"quarter"` // 1-4
	Year       int      `json:"year"`
	Outlook    string   `json:"outlook"`  // "positivo", "neutro", "desafiante"
	KeyEvents  []string `json:"key_events"`
	ActionItems []string `json:"action_items"`
	Summary    string   `json:"summary"`
}

// TeamScore holds synastry-based compatibility for a team member.
type TeamScore struct {
	Name       string   `json:"name"`
	Score      int      `json:"score"`      // 0-100
	Strengths  []string `json:"strengths"`
	Tensions   []string `json:"tensions"`
}

// MercuryRxPeriod holds a Mercury retrograde window.
type MercuryRxPeriod struct {
	StartMonth int    `json:"start_month"`
	StartDay   int    `json:"start_day"`
	EndMonth   int    `json:"end_month"`
	EndDay     int    `json:"end_day"`
	Sign       string `json:"sign"`
	Impact     string `json:"impact"` // business impact description
}

// DashboardResponse is the full business dashboard data.
type DashboardResponse struct {
	CompanyName    string             `json:"company_name"`
	Year           int                `json:"year"`
	Month          int                `json:"month"`
	TodayAgenda    []AgendaItem       `json:"today_agenda"`
	WeekHighlights []string           `json:"week_highlights"`
	CashFlow       []CashFlowMonth    `json:"cash_flow"`
	RiskHeatmap    []RiskCell         `json:"risk_heatmap"`
	TopTimings     []TimingWindow     `json:"top_timings"`
	MercuryRx      []MercuryRxPeriod  `json:"mercury_rx"`
	Forecast       *QuarterlyForecast `json:"quarterly_forecast,omitempty"`
}

// MonthLabels in Spanish.
var MonthLabels = [12]string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

// RiskCategories for the heatmap.
var RiskCategories = []string{"financiero", "operativo", "legal", "personal", "reputacional"}

// Business house mapping — which natal houses map to which business function.
var BusinessHouses = map[string][]int{
	"revenue":    {2},      // House 2: income, liquid assets
	"expenses":   {8},      // House 8: debt, shared resources
	"reputation": {10, 1},  // House 10: public image, House 1: identity
	"operations": {6},      // House 6: daily work, employees
	"legal":      {7, 9},   // House 7: contracts, House 9: legal proceedings
	"personnel":  {6, 11},  // House 6: employees, House 11: allies
	"strategy":   {10, 11}, // House 10: authority, House 11: goals
}
