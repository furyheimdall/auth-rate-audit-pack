package pack

// BaselineSnapshot is the ShopifyQL auth-rate baseline + period/YoY delta
// the pack consumes. Other seats may later satisfy BaselineSource; fixtures
// implement it today so baseline/ can stay stubbed.
//
// Rates are stored as percents (e.g. 93.2). Deltas are percentage points.
type BaselineSnapshot struct {
	Shop            string  `json:"shop"`
	Gateway         string  `json:"gateway"`
	PeriodLabel     string  `json:"period_label"`
	PeriodRate      float64 `json:"period_rate"`
	PriorPeriodRate float64 `json:"prior_period_rate"`
	YoYRate         float64 `json:"yoy_rate"`
	PeriodDeltaPP   float64 `json:"period_delta_pp"`
	YoYDeltaPP      float64 `json:"yoy_delta_pp"`
}

// Recorded reports whether the snapshot has enough fields to score as present.
func (b BaselineSnapshot) Recorded() bool {
	return b.Shop != "" && b.Gateway != "" && b.PeriodLabel != ""
}

// DriftFlag is the >0.5pp unexplained-drop input from the drift seat.
// A flag is a drift signal, not a fraud verdict.
type DriftFlag struct {
	ObservedDeltaPP float64 `json:"observed_delta_pp"`
	ThresholdPP     float64 `json:"threshold_pp"`
	Unexplained     bool    `json:"unexplained"`
	Flagged         bool    `json:"flagged"`
}

// Recorded reports whether a drift result was supplied (vs a zero value).
func (d DriftFlag) Recorded() bool {
	return d.ThresholdPP != 0 || d.ObservedDeltaPP != 0 || d.Flagged || d.Unexplained
}

// InventoryItem is one expected-vs-actual Admin webhook / OrderRiskAssessment row.
// A gap is an inventory finding; inventory ≠ remediation.
type InventoryItem struct {
	Topic    string `json:"topic"`
	Expected bool   `json:"expected"`
	Present  bool   `json:"present"`
}

// Gap is true when the topic is expected but not present.
func (i InventoryItem) Gap() bool {
	return i.Expected && !i.Present
}

// InventoryChecklist is the inventory seat output the pack consumes.
type InventoryChecklist struct {
	Items []InventoryItem `json:"items"`
}

// Gaps returns expected-but-missing items.
func (c InventoryChecklist) Gaps() []InventoryItem {
	var out []InventoryItem
	for _, it := range c.Items {
		if it.Gap() {
			out = append(out, it)
		}
	}
	return out
}

// CheckStatus is a staging pass/fail/skip cell. Checklist only — not full QA.
type CheckStatus string

const (
	StatusPass CheckStatus = "pass"
	StatusFail CheckStatus = "fail"
	StatusSkip CheckStatus = "skip"
)

// StagingRow is one Shop Pay / AP / GP / 3DS2 SCA (or similar) matrix cell.
type StagingRow struct {
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
}

// VendorRow is one Signifyd / NoFraud / Kount re-verify row (session vs legacy).
type VendorRow struct {
	Vendor string      `json:"vendor"`
	Mode   string      `json:"mode"`
	Status CheckStatus `json:"status"`
}

// StagingMatrix is the staging seat output the pack consumes.
type StagingMatrix struct {
	Vendors []VendorRow  `json:"vendors"`
	Checks  []StagingRow `json:"checks"`
}

// Inputs is the assembled snapshot from the four seats.
type Inputs struct {
	Baseline  BaselineSnapshot
	Drift     DriftFlag
	Inventory InventoryChecklist
	Staging   StagingMatrix
}

// BaselineSource supplies a baseline snapshot. Fixtures implement this;
// a future baseline/ type can too without this package depending on it.
type BaselineSource interface {
	BaselineSnapshot() (BaselineSnapshot, error)
}

// DriftSource supplies a drift flag.
type DriftSource interface {
	DriftFlag() (DriftFlag, error)
}

// InventorySource supplies an inventory checklist.
type InventorySource interface {
	InventoryChecklist() (InventoryChecklist, error)
}

// StagingSource supplies a staging matrix.
type StagingSource interface {
	StagingMatrix() (StagingMatrix, error)
}

// Sources is the four seats the pack assembles. Any source may be a fixture.
type Sources struct {
	Baseline  BaselineSource
	Drift     DriftSource
	Inventory InventorySource
	Staging   StagingSource
}
