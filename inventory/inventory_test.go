package inventory

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSmoke(t *testing.T) {
	if Seat != "inventory" {
		t.Fatalf("Seat = %q", Seat)
	}
}

func TestKinds(t *testing.T) {
	if KindWebhook != "admin_webhook" {
		t.Fatalf("KindWebhook = %q", KindWebhook)
	}
	if KindOrderRiskAssessment != "order_risk_assessment" {
		t.Fatalf("KindOrderRiskAssessment = %q", KindOrderRiskAssessment)
	}
}

func TestDefaultExpected(t *testing.T) {
	got := DefaultExpected()
	want := []string{
		TopicOrdersCreate,
		TopicOrdersUpdated,
		TopicOrdersPaid,
		TopicOrdersCancelled,
		TopicOrdersRiskAssessmentChanged,
	}
	if len(got) != len(want) {
		t.Fatalf("DefaultExpected len = %d, want %d", len(got), len(want))
	}
	for i, it := range got {
		if it.Kind != KindWebhook || it.Key != want[i] {
			t.Fatalf("DefaultExpected[%d] = %+v", i, it)
		}
	}
}

func TestCheckFixtureGaps(t *testing.T) {
	expected := loadItems(t, "testdata/expected.json")
	actual := loadItems(t, "testdata/actual_gaps.json")
	cl := Check(expected, actual)

	gaps := cl.Gaps()
	if len(gaps) != 3 {
		t.Fatalf("gaps = %v", gaps)
	}
	assertRow(t, gaps[0], KindWebhook, TopicOrdersCancelled, StatusGap)
	assertRow(t, gaps[1], KindWebhook, TopicOrdersRiskAssessmentChanged, StatusGap)
	assertRow(t, gaps[2], KindOrderRiskAssessment, "signifyd", StatusGap)

	extra := cl.Unexpected()
	if len(extra) != 1 {
		t.Fatalf("unexpected = %v", extra)
	}
	assertRow(t, extra[0], KindWebhook, "APP_UNINSTALLED", StatusUnexpected)
}

func TestCheckFixtureMatch(t *testing.T) {
	expected := loadItems(t, "testdata/expected.json")
	actual := loadItems(t, "testdata/actual_match.json")
	cl := Check(expected, actual)
	if n := len(cl.Gaps()); n != 0 {
		t.Fatalf("gaps = %d", n)
	}
	if n := len(cl.Unexpected()); n != 0 {
		t.Fatalf("unexpected = %d", n)
	}
	if len(cl.Rows) != len(expected) {
		t.Fatalf("rows = %d, want %d", len(cl.Rows), len(expected))
	}
	for _, r := range cl.Rows {
		if r.Status != StatusMatch || !r.Expected || !r.Actual {
			t.Fatalf("row = %+v", r)
		}
	}
}

func TestCheckEmptyActual(t *testing.T) {
	cl := Check(DefaultExpected(), nil)
	if len(cl.Gaps()) != len(DefaultExpected()) {
		t.Fatalf("gaps = %v", cl.Gaps())
	}
	if len(cl.Unexpected()) != 0 {
		t.Fatalf("unexpected = %v", cl.Unexpected())
	}
}

type itemFile struct {
	Items []Item `json:"items"`
}

func loadItems(t *testing.T, path string) []Item {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f itemFile
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Items) == 0 {
		t.Fatalf("empty fixture %s", path)
	}
	return f.Items
}

func assertRow(t *testing.T, r Row, kind Kind, key string, st Status) {
	t.Helper()
	if r.Kind != kind || r.Key != key || r.Status != st {
		t.Fatalf("row = %+v, want kind=%s key=%s status=%s", r, kind, key, st)
	}
}
