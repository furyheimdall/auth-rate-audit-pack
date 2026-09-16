// Package drift flags authorization-rate movement versus a baseline.
//
// Auth rate drift ≠ fraud verdict — a >0.5pp unexplained drop is a drift
// flag, not a fraud decision. Evaluation is pure arithmetic over fixture
// or recorded rates; this package does not call the network.
package drift

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
)

// Seat is the locked package seat name.
const Seat = "drift"

// ThresholdPP is the locked day-1 drift flag: unexplained drop >0.5
// percentage points versus baseline / period / YoY.
const ThresholdPP = 0.5

// KindBaseline is a direct current-versus-baseline comparison (not period/YoY).
const KindBaseline baseline.Kind = "baseline"

// Input is one current-versus-compare rate observation.
// Compare is typically a baseline.Record current rate or a period/YoY prior.
type Input struct {
	Current     baseline.Rate `json:"current_pp"`
	Compare     baseline.Rate `json:"compare_pp"`
	ExplainedPP float64       `json:"explained_pp,omitempty"`
	Kind        baseline.Kind `json:"kind,omitempty"`
}

// Result is a drift evaluation. Flagged is not a fraud verdict.
type Result struct {
	Flagged       bool          `json:"flagged"`
	DropPP        float64       `json:"drop_pp"`
	UnexplainedPP float64       `json:"unexplained_pp"`
	ThresholdPP   float64       `json:"threshold_pp"`
	Kind          baseline.Kind `json:"kind,omitempty"`
	Reason        string        `json:"reason"`
}

// FraudVerdict is always false. A drift flag is not a fraud decision.
func (r Result) FraudVerdict() bool { return false }

// InputFromDelta builds an Input from a baseline period or YoY delta.
func InputFromDelta(d baseline.Delta, explainedPP float64) Input {
	return Input{
		Current:     d.Current,
		Compare:     d.Compare,
		ExplainedPP: explainedPP,
		Kind:        d.Kind,
	}
}

// Evaluate compares current to compare and flags when the unexplained
// drop exceeds ThresholdPP. A flag is a drift signal, not a fraud verdict.
func Evaluate(in Input) (Result, error) {
	if err := validRate(in.Current, "current_pp"); err != nil {
		return Result{}, err
	}
	if err := validRate(in.Compare, "compare_pp"); err != nil {
		return Result{}, err
	}
	if in.ExplainedPP < 0 {
		return Result{}, fmt.Errorf("drift: explained_pp must be >= 0")
	}

	drop := float64(in.Compare - in.Current)
	if drop < 0 {
		drop = 0
	}
	unexplained := drop - in.ExplainedPP
	if unexplained < 0 {
		unexplained = 0
	}

	out := Result{
		Flagged:       unexplained > ThresholdPP,
		DropPP:        drop,
		UnexplainedPP: unexplained,
		ThresholdPP:   ThresholdPP,
		Kind:          in.Kind,
	}
	switch {
	case out.Flagged:
		out.Reason = fmt.Sprintf("unexplained drop %.3fpp exceeds %.1fpp (%s); not a fraud verdict", unexplained, ThresholdPP, kindOrBaseline(in.Kind))
	case drop == 0:
		out.Reason = "no authorization-rate drop"
	default:
		out.Reason = fmt.Sprintf("unexplained drop %.3fpp does not exceed %.1fpp; not a fraud verdict", unexplained, ThresholdPP)
	}
	return out, nil
}

// Flag reports whether Evaluate(in) would flag. A flag is not a fraud verdict.
func Flag(in Input) (bool, error) {
	res, err := Evaluate(in)
	if err != nil {
		return false, err
	}
	return res.Flagged, nil
}

// EvaluateRecord flags when any recorded period or YoY unexplained drop
// exceeds ThresholdPP. Period/YoY moves are drift inputs, not verdicts.
func EvaluateRecord(rec baseline.Record, explainedPP float64) (Result, error) {
	if rec.Period == nil && rec.YoY == nil {
		return Result{}, fmt.Errorf("drift: record has no period or yoy delta")
	}
	var best Result
	var found bool
	for _, d := range []*baseline.Delta{rec.Period, rec.YoY} {
		if d == nil {
			continue
		}
		res, err := Evaluate(InputFromDelta(*d, explainedPP))
		if err != nil {
			return Result{}, err
		}
		if !found || res.UnexplainedPP > best.UnexplainedPP {
			best = res
			found = true
		}
	}
	return best, nil
}

// Parse decodes a single Input fixture.
func Parse(data []byte) (Input, error) {
	var in Input
	if err := json.Unmarshal(data, &in); err != nil {
		return Input{}, fmt.Errorf("drift: parse: %w", err)
	}
	return in, nil
}

// Load reads an Input JSON file (typically testdata).
func Load(path string) (Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Input{}, fmt.Errorf("drift: load %s: %w", path, err)
	}
	return Parse(data)
}

func validRate(r baseline.Rate, field string) error {
	if r < 0 || r > 100 {
		return fmt.Errorf("drift: %s must be in [0, 100], got %v", field, r)
	}
	return nil
}

func kindOrBaseline(k baseline.Kind) string {
	if k == "" {
		return string(KindBaseline)
	}
	return string(k)
}
