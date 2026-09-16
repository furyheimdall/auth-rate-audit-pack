package baseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSmoke(t *testing.T) {
	if Seat != "baseline" {
		t.Fatalf("Seat = %q", Seat)
	}
	if Metric != "payment_authorization_rate" {
		t.Fatalf("Metric = %q", Metric)
	}
	if !strings.Contains(Query, Metric) || !strings.Contains(Query, "FROM payment_attempts") {
		t.Fatalf("Query = %q", Query)
	}
	if !strings.Contains(QueryYoY, "COMPARE TO last_year") {
		t.Fatalf("QueryYoY = %q", QueryYoY)
	}
}

func TestFromCounts(t *testing.T) {
	r, err := FromCounts(9425, 10000)
	if err != nil {
		t.Fatal(err)
	}
	if r != 94.25 {
		t.Fatalf("Rate = %v", r)
	}
	if r.PercentagePoints() != 94.25 {
		t.Fatalf("PercentagePoints = %v", r.PercentagePoints())
	}
	if _, err := FromCounts(1, 0); err == nil {
		t.Fatal("expected unique=0 error")
	}
	if _, err := FromCounts(3, 2); err == nil {
		t.Fatal("expected successful>unique error")
	}
}

func TestLoadShopifyQLFixture(t *testing.T) {
	rec, err := Load(filepath.Join("testdata", "shopifyql_payment_authorization_rate.json"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Metric != Metric {
		t.Fatalf("metric = %q", rec.Metric)
	}
	if rec.Current.Rate != 94.25 {
		t.Fatalf("current rate = %v", rec.Current.Rate)
	}
	if rec.Period == nil || rec.YoY == nil {
		t.Fatal("expected period and yoy deltas")
	}
	if rec.Period.Kind != KindPeriod || rec.YoY.Kind != KindYoY {
		t.Fatalf("kinds = %s %s", rec.Period.Kind, rec.YoY.Kind)
	}
	// 94.25 − 94.50 = −0.25pp period drop
	if rec.Period.DeltaPP != -0.25 || rec.Period.DropPP() != 0.25 {
		t.Fatalf("period delta = %+v", rec.Period)
	}
	// 94.25 − 95.00 = −0.75pp YoY drop
	if rec.YoY.DeltaPP != -0.75 || rec.YoY.DropPP() != 0.75 {
		t.Fatalf("yoy delta = %+v", rec.YoY)
	}
	if rec.Current.Period.Label != "2026-08" || rec.Period.ComparePeriod.Label != "2026-07" {
		t.Fatalf("periods = %+v %+v", rec.Current.Period, rec.Period.ComparePeriod)
	}
}

func TestLoadRateOnlyFixture(t *testing.T) {
	rec, err := Load(filepath.Join("testdata", "rate_only.json"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.Current.Rate != 94.0 {
		t.Fatalf("current = %v", rec.Current.Rate)
	}
	if rec.Period == nil || rec.Period.DeltaPP != 0 {
		t.Fatalf("period = %+v", rec.Period)
	}
	if rec.YoY == nil || rec.YoY.DeltaPP != 0.5 || rec.YoY.DropPP() != 0 {
		t.Fatalf("yoy = %+v", rec.YoY)
	}
}

func TestNewRecordAndDelta(t *testing.T) {
	cur := Snapshot{Rate: 94, Period: Period{Label: "now"}}
	prior := Snapshot{Rate: 95, Period: Period{Label: "prior"}}
	yoy := Snapshot{Rate: 93, Period: Period{Label: "last-year"}}
	rec, err := NewRecord(cur, &prior, &yoy)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Period.DropPP() != 1 || rec.YoY.DropPP() != 0 {
		t.Fatalf("drops period=%v yoy=%v", rec.Period.DropPP(), rec.YoY.DropPP())
	}
	if _, err := NewDelta("week", cur, prior); err == nil {
		t.Fatal("expected unknown kind")
	}
}

func TestSnapshotValidation(t *testing.T) {
	if err := (Snapshot{Rate: -1}).Valid(); err == nil {
		t.Fatal("expected negative rate error")
	}
	if err := (Snapshot{Rate: 101}).Valid(); err == nil {
		t.Fatal("expected >100 rate error")
	}
	if err := (Snapshot{Rate: 50, Period: Period{Start: "2026-13-01"}}).Valid(); err == nil {
		t.Fatal("expected bad date")
	}
	if err := (Snapshot{Rate: 50, Period: Period{Start: "2026-08-31", End: "2026-08-01"}}).Valid(); err == nil {
		t.Fatal("expected inverted period")
	}
	if err := (Snapshot{SuccessfulPayments: 10, UniqueAttempts: 100, Rate: 50}).Valid(); err == nil {
		t.Fatal("expected count/rate mismatch")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := Parse([]byte(`{`)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestREADMESeatCanon(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, row := range []string{
		"| [`baseline/`](baseline/) | ShopifyQL auth-rate baseline + period/YoY delta |",
		"| [`drift/`](drift/) | Flag e.g. >0.5pp unexplained drop |",
	} {
		if !strings.Contains(body, row) {
			t.Fatalf("README seat table drifted from locked canon; missing %q", row)
		}
	}
}
