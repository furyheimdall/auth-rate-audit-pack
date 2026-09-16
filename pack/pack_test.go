package pack

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
	"github.com/furyheimdall/auth-rate-audit-pack/drift"
	"github.com/furyheimdall/auth-rate-audit-pack/inventory"
	"github.com/furyheimdall/auth-rate-audit-pack/staging"
)

func TestSmoke(t *testing.T) {
	if Seat != "pack" {
		t.Fatalf("Seat = %q", Seat)
	}
}

func TestSeats(t *testing.T) {
	got := Seats()
	want := []string{baseline.Seat, drift.Seat, inventory.Seat, staging.Seat, Seat}
	if len(got) != len(want) {
		t.Fatalf("Seats() = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Seats()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAssembleHealthyFixture(t *testing.T) {
	r := mustAssemble(t, "healthy")
	if r.Overall != 100 {
		t.Fatalf("overall = %d, want 100", r.Overall)
	}
	for _, s := range r.Sections {
		if s.Score != 100 {
			t.Fatalf("section %s = %d, want 100", s.Seat, s.Score)
		}
		if s.Weight != 25 {
			t.Fatalf("section %s weight = %d", s.Seat, s.Weight)
		}
	}
	if r.Inputs.Baseline.Gateway != "Adyen" {
		t.Fatalf("gateway = %q", r.Inputs.Baseline.Gateway)
	}
	if r.Inputs.Drift.Flagged {
		t.Fatal("healthy fixture should not flag drift")
	}
}

func TestAssembleDriftFixture(t *testing.T) {
	r := mustAssemble(t, "drift")
	// baseline 100, drift 40, inventory 2/4 = 50, staging 4 pass / 1 fail = 80
	// (100*25 + 40*25 + 50*25 + 80*25 + 50) / 100 = 68
	if r.Overall != 68 {
		t.Fatalf("overall = %d, want 68", r.Overall)
	}
	bySeat := map[string]int{}
	for _, s := range r.Sections {
		bySeat[s.Seat] = s.Score
	}
	if bySeat[baseline.Seat] != 100 || bySeat[drift.Seat] != 40 || bySeat[inventory.Seat] != 50 || bySeat[staging.Seat] != 80 {
		t.Fatalf("sections = %v", bySeat)
	}
	if !r.Inputs.Drift.Flagged || !r.Inputs.Drift.Unexplained {
		t.Fatalf("drift = %+v", r.Inputs.Drift)
	}
	if got := len(r.Inputs.Inventory.Gaps()); got != 2 {
		t.Fatalf("gaps = %d, want 2", got)
	}
}

func TestMarkdownCanon(t *testing.T) {
	md := mustAssemble(t, "drift").Markdown()
	for _, needle := range []string{
		"**Overall score:** 68 / 100",
		Audience,
		ICP,
		SeasonNote,
		AnchorAuthRate,
		AnchorWebhooks,
		OUT,
		"not a fraud verdict",
		"inventory ≠ remediation",
		"checklist, not full QA",
		"plus-dtc-stripe-fixture",
		"OrderRiskAssessment",
		"Shop Pay",
		"Signifyd",
		"NoFraud",
		"Kount",
		"Adyen/Braintree/Stripe",
		"agency payments lead",
		"Pre-BFCM / Q4",
	} {
		if !strings.Contains(md, needle) {
			t.Fatalf("markdown missing %q\n%s", needle, md)
		}
	}
	if strings.Contains(md, "SaaS") {
		t.Fatal("thin docs/report must not grow a SaaS landing page")
	}
}

func TestPDFStub(t *testing.T) {
	var buf bytes.Buffer
	if err := mustAssemble(t, "healthy").WritePDF(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.Bytes()
	if !bytes.HasPrefix(got, []byte("%PDF-1.4")) {
		t.Fatalf("pdf prefix = %q", got[:min(20, len(got))])
	}
	if !bytes.Contains(got, []byte("%%EOF")) {
		t.Fatal("pdf missing EOF")
	}
	if !bytes.Contains(got, []byte("Overall score: 100 / 100")) {
		t.Fatalf("pdf missing score: %s", got)
	}
}

func TestWriteFormats(t *testing.T) {
	r := mustAssemble(t, "healthy")
	var md, pdf bytes.Buffer
	if err := r.Write(&md, FormatMarkdown); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md.String(), "# Auth Rate Audit Pack") {
		t.Fatalf("markdown = %s", md.String())
	}
	if err := r.Write(&pdf, FormatPDF); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf.Bytes(), []byte("%PDF")) {
		t.Fatal("pdf write")
	}
	if err := r.Write(&md, Format("html")); err == nil {
		t.Fatal("unknown format should error")
	}
}

func TestCollectRequiresSources(t *testing.T) {
	if _, err := Collect(Sources{}); err == nil {
		t.Fatal("expected error for nil sources")
	}
	f := mustFixture(t, "healthy")
	if _, err := Collect(Sources{Baseline: f}); err == nil {
		t.Fatal("expected error when drift/inventory/staging missing")
	}
}

func TestShopifyPaymentsNotPrimaryICP(t *testing.T) {
	f := mustFixture(t, "healthy")
	f.Baseline.Gateway = "Shopify Payments"
	r, err := Assemble(f.Sources())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, fn := range r.Findings {
		if strings.Contains(fn.Text, "Shopify Payments-native is out") {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %+v", r.Findings)
	}
}

func TestMissingBaselineScoresZero(t *testing.T) {
	r := Score(Inputs{})
	if r.Overall != 0 {
		t.Fatalf("empty inputs overall = %d", r.Overall)
	}
	for _, s := range r.Sections {
		if s.Score != 0 {
			t.Fatalf("%s = %d", s.Seat, s.Score)
		}
	}
}

func TestLoadFixtureUnknown(t *testing.T) {
	_, err := LoadFixture(FixtureFS(), "not-a-fixture")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "healthy") || !strings.Contains(err.Error(), "drift") {
		t.Fatalf("error should list fixtures: %v", err)
	}
}

func TestFixtureNames(t *testing.T) {
	names, err := FixtureNames(FixtureFS())
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "healthy") || !strings.Contains(joined, "drift") {
		t.Fatalf("names = %v", names)
	}
}

func TestDirFSTestdata(t *testing.T) {
	// Same fixtures via the on-disk testdata dir (still no network).
	f, err := LoadFixture(os.DirFS("testdata"), "healthy")
	if err != nil {
		t.Fatal(err)
	}
	if f.Baseline.Shop != "plus-dtc-adyen-fixture" {
		t.Fatalf("shop = %q", f.Baseline.Shop)
	}
}

func TestDriftThresholdDefault(t *testing.T) {
	f := mustFixture(t, "healthy")
	f.Drift.ThresholdPP = 0
	in, err := Collect(f.Sources())
	if err != nil {
		t.Fatal(err)
	}
	if in.Drift.ThresholdPP != drift.ThresholdPP {
		t.Fatalf("threshold = %v, want %v", in.Drift.ThresholdPP, drift.ThresholdPP)
	}
}

func TestExplainedFlagScore(t *testing.T) {
	in := mustAssemble(t, "drift").Inputs
	in.Drift.Unexplained = false
	in.Drift.Flagged = true
	r := Score(in)
	for _, s := range r.Sections {
		if s.Seat == drift.Seat && s.Score != 70 {
			t.Fatalf("explained flag score = %d, want 70", s.Score)
		}
	}
}

func mustFixture(t *testing.T, name string) Fixture {
	t.Helper()
	f, err := LoadFixture(FixtureFS(), name)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func mustAssemble(t *testing.T, name string) Report {
	t.Helper()
	r, err := Assemble(mustFixture(t, name).Sources())
	if err != nil {
		t.Fatal(err)
	}
	return r
}
