// Package staging is the vendor re-verify + staging pass/fail seat.
//
// Runbook + matrix template only; not production cutover.
package staging

// Seat is the locked package seat name.
const Seat = "staging"

// Vendor is a fraud vendor in the re-verify runbook.
type Vendor string

const (
	VendorSignifyd Vendor = "signifyd"
	VendorNoFraud  Vendor = "nofraud"
	VendorKount    Vendor = "kount"
)

// Mode is session (modern) vs legacy integration.
type Mode string

const (
	ModeSession Mode = "session"
	ModeLegacy  Mode = "legacy"
)

// StepStatus is one runbook observation. Not a live vendor-console call.
type StepStatus string

const (
	StepVerified StepStatus = "verified"
	StepGap      StepStatus = "gap"
	StepSkip     StepStatus = "skip"
	StepPending  StepStatus = "pending"
)

// RunbookStep is one fraud-vendor re-verify checklist line.
type RunbookStep struct {
	ID     string `json:"id"`
	Vendor Vendor `json:"vendor"`
	Mode   Mode   `json:"mode"`
	Title  string `json:"title"`
}

// Runbook is the Signifyd/NoFraud/Kount session-vs-legacy template.
type Runbook struct {
	Steps []RunbookStep `json:"steps"`
}

// StepObservation is a fixture-backed result for one runbook step.
type StepObservation struct {
	ID     string     `json:"id"`
	Status StepStatus `json:"status"`
	Notes  string     `json:"notes,omitempty"`
}

// FilledStep is a template step plus fixture status.
type FilledStep struct {
	RunbookStep
	Status StepStatus `json:"status"`
	Notes  string     `json:"notes,omitempty"`
}

// FilledRunbook is a completed (or partial) re-verify checklist.
type FilledRunbook struct {
	Steps []FilledStep `json:"steps"`
}

// Wallet is a checkout wallet in the staging matrix.
type Wallet string

const (
	WalletShopPay   Wallet = "shop_pay"
	WalletApplePay  Wallet = "apple_pay"
	WalletGooglePay Wallet = "google_pay"
)

// Result is a matrix cell outcome — checklist, not full QA.
type Result string

const (
	ResultPass    Result = "pass"
	ResultFail    Result = "fail"
	ResultPending Result = "pending"
)

// MatrixCell is one Shop Pay / Apple Pay / Google Pay + optional 3DS2 SCA smoke row.
type MatrixCell struct {
	ID          string `json:"id"`
	Wallet      Wallet `json:"wallet"`
	ThreeDS2SCA bool   `json:"three_ds2_sca"`
	Title       string `json:"title"`
}

// Matrix is the staging pass/fail template (smoke checklist, not full QA).
type Matrix struct {
	Cells []MatrixCell `json:"cells"`
}

// CellObservation is a fixture-backed result for one matrix cell.
type CellObservation struct {
	ID     string `json:"id"`
	Result Result `json:"result"`
	Notes  string `json:"notes,omitempty"`
}

// FilledCell is a template cell plus fixture result.
type FilledCell struct {
	MatrixCell
	Result Result `json:"result"`
	Notes  string `json:"notes,omitempty"`
}

// FilledMatrix is a completed (or partial) staging pass/fail checklist.
type FilledMatrix struct {
	Cells []FilledCell `json:"cells"`
}

// DefaultRunbook is the locked vendor re-verify template
// (Signifyd / NoFraud / Kount — session vs legacy). Not production cutover.
func DefaultRunbook() Runbook {
	return Runbook{Steps: []RunbookStep{
		{ID: "signifyd.session", Vendor: VendorSignifyd, Mode: ModeSession, Title: "Signifyd session / checkout-session path present"},
		{ID: "signifyd.legacy", Vendor: VendorSignifyd, Mode: ModeLegacy, Title: "Signifyd legacy plugin / fingerprint path documented, not mixed into session"},
		{ID: "nofraud.session", Vendor: VendorNoFraud, Mode: ModeSession, Title: "NoFraud session / checkout-session path present"},
		{ID: "nofraud.legacy", Vendor: VendorNoFraud, Mode: ModeLegacy, Title: "NoFraud legacy script / webhook path documented, not mixed into session"},
		{ID: "kount.session", Vendor: VendorKount, Mode: ModeSession, Title: "Kount session path present"},
		{ID: "kount.legacy", Vendor: VendorKount, Mode: ModeLegacy, Title: "Kount legacy RIS / device-data path documented, not mixed into session"},
	}}
}

// ApplyRunbook overlays fixture observations onto the runbook template.
// Missing observations stay pending. No live vendor consoles.
func ApplyRunbook(rb Runbook, observed []StepObservation) FilledRunbook {
	byID := make(map[string]StepObservation, len(observed))
	for _, o := range observed {
		byID[o.ID] = o
	}
	out := FilledRunbook{Steps: make([]FilledStep, 0, len(rb.Steps))}
	for _, s := range rb.Steps {
		fs := FilledStep{RunbookStep: s, Status: StepPending}
		if o, ok := byID[s.ID]; ok {
			fs.Status = o.Status
			fs.Notes = o.Notes
		}
		out.Steps = append(out.Steps, fs)
	}
	return out
}

// Gaps returns runbook steps observed as gaps.
func (r FilledRunbook) Gaps() []FilledStep {
	out := make([]FilledStep, 0)
	for _, s := range r.Steps {
		if s.Status == StepGap {
			out = append(out, s)
		}
	}
	return out
}

// DefaultMatrix is the Shop Pay / Apple Pay / Google Pay + 3DS2 SCA smoke
// template. Checklist only — not full QA, not production cutover.
func DefaultMatrix() Matrix {
	return Matrix{Cells: []MatrixCell{
		{ID: "shop_pay", Wallet: WalletShopPay, ThreeDS2SCA: false, Title: "Shop Pay wallet smoke"},
		{ID: "apple_pay", Wallet: WalletApplePay, ThreeDS2SCA: false, Title: "Apple Pay wallet smoke"},
		{ID: "google_pay", Wallet: WalletGooglePay, ThreeDS2SCA: false, Title: "Google Pay wallet smoke"},
		{ID: "shop_pay.3ds2_sca", Wallet: WalletShopPay, ThreeDS2SCA: true, Title: "Shop Pay 3DS2 SCA smoke"},
		{ID: "apple_pay.3ds2_sca", Wallet: WalletApplePay, ThreeDS2SCA: true, Title: "Apple Pay 3DS2 SCA smoke"},
		{ID: "google_pay.3ds2_sca", Wallet: WalletGooglePay, ThreeDS2SCA: true, Title: "Google Pay 3DS2 SCA smoke"},
	}}
}

// ApplyMatrix overlays fixture results onto the matrix template.
// Missing observations stay pending. No live staging shops.
func ApplyMatrix(m Matrix, observed []CellObservation) FilledMatrix {
	byID := make(map[string]CellObservation, len(observed))
	for _, o := range observed {
		byID[o.ID] = o
	}
	out := FilledMatrix{Cells: make([]FilledCell, 0, len(m.Cells))}
	for _, c := range m.Cells {
		fc := FilledCell{MatrixCell: c, Result: ResultPending}
		if o, ok := byID[c.ID]; ok {
			fc.Result = o.Result
			fc.Notes = o.Notes
		}
		out.Cells = append(out.Cells, fc)
	}
	return out
}

// Failed returns matrix cells observed as fail.
func (m FilledMatrix) Failed() []FilledCell {
	out := make([]FilledCell, 0)
	for _, c := range m.Cells {
		if c.Result == ResultFail {
			out = append(out, c)
		}
	}
	return out
}

// AllPassed reports whether every template cell is pass.
func (m FilledMatrix) AllPassed() bool {
	if len(m.Cells) == 0 {
		return false
	}
	for _, c := range m.Cells {
		if c.Result != ResultPass {
			return false
		}
	}
	return true
}
