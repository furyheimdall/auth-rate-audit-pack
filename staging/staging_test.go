package staging

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSmoke(t *testing.T) {
	if Seat != "staging" {
		t.Fatalf("Seat = %q", Seat)
	}
}

func TestDefaultRunbook(t *testing.T) {
	rb := DefaultRunbook()
	if len(rb.Steps) != 6 {
		t.Fatalf("steps = %d", len(rb.Steps))
	}
	want := []struct {
		id     string
		vendor Vendor
		mode   Mode
	}{
		{"signifyd.session", VendorSignifyd, ModeSession},
		{"signifyd.legacy", VendorSignifyd, ModeLegacy},
		{"nofraud.session", VendorNoFraud, ModeSession},
		{"nofraud.legacy", VendorNoFraud, ModeLegacy},
		{"kount.session", VendorKount, ModeSession},
		{"kount.legacy", VendorKount, ModeLegacy},
	}
	for i, w := range want {
		s := rb.Steps[i]
		if s.ID != w.id || s.Vendor != w.vendor || s.Mode != w.mode || s.Title == "" {
			t.Fatalf("step[%d] = %+v, want %+v", i, s, w)
		}
	}
}

func TestApplyRunbookVerifiedFixture(t *testing.T) {
	obs := loadSteps(t, "testdata/runbook_verified.json")
	filled := ApplyRunbook(DefaultRunbook(), obs)
	if n := len(filled.Gaps()); n != 0 {
		t.Fatalf("gaps = %d", n)
	}
	if filled.Steps[0].Status != StepVerified {
		t.Fatalf("signifyd.session = %s", filled.Steps[0].Status)
	}
	if filled.Steps[2].Status != StepSkip {
		t.Fatalf("nofraud.session = %s", filled.Steps[2].Status)
	}
}

func TestApplyRunbookSessionGapFixture(t *testing.T) {
	obs := loadSteps(t, "testdata/runbook_session_gap.json")
	filled := ApplyRunbook(DefaultRunbook(), obs)
	gaps := filled.Gaps()
	if len(gaps) != 1 {
		t.Fatalf("gaps = %v", gaps)
	}
	if gaps[0].ID != "signifyd.session" || gaps[0].Mode != ModeSession {
		t.Fatalf("gap = %+v", gaps[0])
	}
}

func TestApplyRunbookPendingWhenMissing(t *testing.T) {
	filled := ApplyRunbook(DefaultRunbook(), nil)
	if len(filled.Steps) != 6 {
		t.Fatalf("steps = %d", len(filled.Steps))
	}
	for _, s := range filled.Steps {
		if s.Status != StepPending {
			t.Fatalf("step %s = %s", s.ID, s.Status)
		}
	}
}

func TestDefaultMatrix(t *testing.T) {
	m := DefaultMatrix()
	if len(m.Cells) != 6 {
		t.Fatalf("cells = %d", len(m.Cells))
	}
	wallets := map[Wallet]int{}
	sca := 0
	for _, c := range m.Cells {
		if c.ID == "" || c.Title == "" {
			t.Fatalf("empty cell %+v", c)
		}
		wallets[c.Wallet]++
		if c.ThreeDS2SCA {
			sca++
		}
	}
	if wallets[WalletShopPay] != 2 || wallets[WalletApplePay] != 2 || wallets[WalletGooglePay] != 2 {
		t.Fatalf("wallets = %v", wallets)
	}
	if sca != 3 {
		t.Fatalf("3DS2 SCA rows = %d", sca)
	}
}

func TestApplyMatrixPassFixture(t *testing.T) {
	obs := loadCells(t, "testdata/matrix_pass.json")
	filled := ApplyMatrix(DefaultMatrix(), obs)
	if !filled.AllPassed() {
		t.Fatalf("want all pass, got %+v", filled.Cells)
	}
	if n := len(filled.Failed()); n != 0 {
		t.Fatalf("failed = %d", n)
	}
}

func TestApplyMatrixFailFixture(t *testing.T) {
	obs := loadCells(t, "testdata/matrix_fail.json")
	filled := ApplyMatrix(DefaultMatrix(), obs)
	if filled.AllPassed() {
		t.Fatal("AllPassed = true")
	}
	failed := filled.Failed()
	if len(failed) != 2 {
		t.Fatalf("failed = %v", failed)
	}
	if failed[0].ID != "google_pay" || failed[1].ID != "google_pay.3ds2_sca" {
		t.Fatalf("failed ids = %s, %s", failed[0].ID, failed[1].ID)
	}
	if filled.Cells[4].Result != ResultPending {
		t.Fatalf("apple_pay.3ds2_sca = %s", filled.Cells[4].Result)
	}
}

func TestApplyMatrixPendingWhenMissing(t *testing.T) {
	filled := ApplyMatrix(DefaultMatrix(), nil)
	if filled.AllPassed() {
		t.Fatal("empty observations must not pass")
	}
	for _, c := range filled.Cells {
		if c.Result != ResultPending {
			t.Fatalf("cell %s = %s", c.ID, c.Result)
		}
	}
}

type stepFile struct {
	Steps []StepObservation `json:"steps"`
}

func loadSteps(t *testing.T, path string) []StepObservation {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f stepFile
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Steps) == 0 {
		t.Fatalf("empty fixture %s", path)
	}
	return f.Steps
}

type cellFile struct {
	Cells []CellObservation `json:"cells"`
}

func loadCells(t *testing.T, path string) []CellObservation {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f cellFile
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cells) == 0 {
		t.Fatalf("empty fixture %s", path)
	}
	return f.Cells
}
