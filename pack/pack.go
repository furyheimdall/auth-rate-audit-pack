// Package pack assembles baseline + drift + inventory + staging into one
// scored report for agency SOW attach.
//
// The report is a scored checklist, not a fraud verdict, not a gateway, and
// not a full fraud platform. Other seats may stay stubbed: this package
// assembles from BaselineSource / DriftSource / InventorySource /
// StagingSource interfaces and fixtures (no network).
package pack

import (
	"fmt"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
	"github.com/furyheimdall/auth-rate-audit-pack/drift"
	"github.com/furyheimdall/auth-rate-audit-pack/inventory"
	"github.com/furyheimdall/auth-rate-audit-pack/staging"
)

// Seat is the locked package seat name.
const Seat = "pack"

// Locked seat names the report sections must cover.
var seats = []string{
	baseline.Seat,
	drift.Seat,
	inventory.Seat,
	staging.Seat,
	Seat,
}

// Seats returns the locked seat names in assembly order, ending with pack.
func Seats() []string {
	out := make([]string, len(seats))
	copy(out, seats)
	return out
}

// Collect materializes inputs from the four seat sources. A nil source is an
// error — pass a fixture if the implementing seat is still a stub.
func Collect(src Sources) (Inputs, error) {
	if src.Baseline == nil {
		return Inputs{}, fmt.Errorf("pack: %s source is required (use a fixture if the seat is still a stub)", baseline.Seat)
	}
	if src.Drift == nil {
		return Inputs{}, fmt.Errorf("pack: %s source is required (use a fixture if the seat is still a stub)", drift.Seat)
	}
	if src.Inventory == nil {
		return Inputs{}, fmt.Errorf("pack: %s source is required (use a fixture if the seat is still a stub)", inventory.Seat)
	}
	if src.Staging == nil {
		return Inputs{}, fmt.Errorf("pack: %s source is required (use a fixture if the seat is still a stub)", staging.Seat)
	}

	b, err := src.Baseline.BaselineSnapshot()
	if err != nil {
		return Inputs{}, fmt.Errorf("pack: %s: %w", baseline.Seat, err)
	}
	d, err := src.Drift.DriftFlag()
	if err != nil {
		return Inputs{}, fmt.Errorf("pack: %s: %w", drift.Seat, err)
	}
	if d.ThresholdPP == 0 {
		d.ThresholdPP = drift.ThresholdPP
	}
	inv, err := src.Inventory.InventoryChecklist()
	if err != nil {
		return Inputs{}, fmt.Errorf("pack: %s: %w", inventory.Seat, err)
	}
	st, err := src.Staging.StagingMatrix()
	if err != nil {
		return Inputs{}, fmt.Errorf("pack: %s: %w", staging.Seat, err)
	}
	return Inputs{Baseline: b, Drift: d, Inventory: inv, Staging: st}, nil
}

// Assemble collects seat inputs and scores them into one agency report.
func Assemble(src Sources) (Report, error) {
	in, err := Collect(src)
	if err != nil {
		return Report{}, err
	}
	return Score(in), nil
}
