package drift

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
)

func TestSmoke(t *testing.T) {
	if Seat != "drift" {
		t.Fatalf("Seat = %q", Seat)
	}
	if ThresholdPP != 0.5 {
		t.Fatalf("ThresholdPP = %v", ThresholdPP)
	}
}

type fixtureCase struct {
	Name              string  `json:"name"`
	CurrentPP         float64 `json:"current_pp"`
	ComparePP         float64 `json:"compare_pp"`
	ExplainedPP       float64 `json:"explained_pp"`
	Kind              string  `json:"kind"`
	WantFlag          bool    `json:"want_flag"`
	WantDropPP        float64 `json:"want_drop_pp"`
	WantUnexplainedPP float64 `json:"want_unexplained_pp"`
}

func TestEvaluateFixtureCases(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wrap struct {
		Cases []fixtureCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		t.Fatal(err)
	}
	if len(wrap.Cases) < 6 {
		t.Fatalf("need flag and no-flag cases, got %d", len(wrap.Cases))
	}

	var sawFlag, sawNoFlag bool
	for _, tc := range wrap.Cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			in := Input{
				Current:     baseline.Rate(tc.CurrentPP),
				Compare:     baseline.Rate(tc.ComparePP),
				ExplainedPP: tc.ExplainedPP,
				Kind:        baseline.Kind(tc.Kind),
			}
			res, err := Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			flagged, err := Flag(in)
			if err != nil {
				t.Fatal(err)
			}
			if res.Flagged != tc.WantFlag || flagged != tc.WantFlag {
				t.Fatalf("flagged=%v flag=%v want %v (%s)", res.Flagged, flagged, tc.WantFlag, res.Reason)
			}
			if res.FraudVerdict() {
				t.Fatal("drift flag must not be a fraud verdict")
			}
			if res.ThresholdPP != ThresholdPP {
				t.Fatalf("threshold = %v", res.ThresholdPP)
			}
			if !almost(res.DropPP, tc.WantDropPP) || !almost(res.UnexplainedPP, tc.WantUnexplainedPP) {
				t.Fatalf("drop=%v unexplained=%v want %v / %v", res.DropPP, res.UnexplainedPP, tc.WantDropPP, tc.WantUnexplainedPP)
			}
			if !strings.Contains(res.Reason, "not a fraud verdict") && res.DropPP > 0 {
				t.Fatalf("reason should restate the anchor: %q", res.Reason)
			}
			if tc.WantFlag {
				sawFlag = true
			} else {
				sawNoFlag = true
			}
		})
	}
	if !sawFlag || !sawNoFlag {
		t.Fatal("fixture cases must cover both flag and no-flag")
	}
}

func TestEvaluateRecordFixtures(t *testing.T) {
	flagRec, err := baseline.Load(filepath.Join("testdata", "record_flag.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := EvaluateRecord(flagRec, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Flagged || res.Kind != baseline.KindPeriod {
		t.Fatalf("flag record: %+v", res)
	}
	if res.FraudVerdict() {
		t.Fatal("EvaluateRecord flag is not a fraud verdict")
	}

	// The same 0.75pp explanation covers the period drop; leftover YoY drop is 0.
	explained, err := EvaluateRecord(flagRec, 0.75)
	if err != nil {
		t.Fatal(err)
	}
	if explained.Flagged {
		t.Fatalf("explained period drop should not flag: %+v", explained)
	}

	noFlag, err := baseline.Load(filepath.Join("testdata", "record_no_flag.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err = EvaluateRecord(noFlag, 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Flagged {
		t.Fatalf("no-flag record flagged: %+v", res)
	}
}

func TestInputFromDelta(t *testing.T) {
	cur := baseline.Snapshot{Rate: 93.25, Period: baseline.Period{Label: "now"}}
	prior := baseline.Snapshot{Rate: 94, Period: baseline.Period{Label: "prior"}}
	d, err := baseline.NewDelta(baseline.KindYoY, cur, prior)
	if err != nil {
		t.Fatal(err)
	}
	in := InputFromDelta(d, 0.1)
	if in.Kind != baseline.KindYoY || in.ExplainedPP != 0.1 || in.Current != 93.25 || in.Compare != 94 {
		t.Fatalf("input = %+v", in)
	}
	flagged, err := Flag(in)
	if err != nil {
		t.Fatal(err)
	}
	if !flagged {
		t.Fatal("expected yoy unexplained drop to flag")
	}
}

func TestEvaluateErrors(t *testing.T) {
	if _, err := Evaluate(Input{Current: -1, Compare: 90}); err == nil {
		t.Fatal("expected current rate error")
	}
	if _, err := Evaluate(Input{Current: 90, Compare: 101}); err == nil {
		t.Fatal("expected compare rate error")
	}
	if _, err := Evaluate(Input{Current: 90, Compare: 91, ExplainedPP: -0.1}); err == nil {
		t.Fatal("expected explained_pp error")
	}
	if _, err := EvaluateRecord(baseline.Record{}, 0); err == nil {
		t.Fatal("expected empty record error")
	}
}

func TestLoadInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "in.json")
	if err := os.WriteFile(path, []byte(`{"current_pp":93.25,"compare_pp":94.0,"kind":"baseline"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	in, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Evaluate(in)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Flagged {
		t.Fatalf("expected flag: %+v", res)
	}
	if _, err := Parse([]byte(`{`)); err == nil {
		t.Fatal("expected parse error")
	}
}

func almost(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}
