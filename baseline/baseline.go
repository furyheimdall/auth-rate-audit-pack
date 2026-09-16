// Package baseline records a ShopifyQL payment_authorization_rate baseline
// plus period and year-over-year deltas.
//
// Auth rate drift ≠ fraud verdict — a period/YoY move is a drift input, not
// a fraud decision. Day-1 collection is fixture-backed; this package does
// not call Admin or execute ShopifyQL.
package baseline

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// Seat is the locked package seat name.
const Seat = "baseline"

// Metric is the ShopifyQL payment_authorization_rate measure.
// Rate = successful_payments / unique_payment_attempts (retries once).
const Metric = "payment_authorization_rate"

// Query documents the day-1 ShopifyQL used to collect the metric.
// Tests never execute this against Admin.
const Query = "" +
	"FROM payment_attempts\n" +
	"  SHOW payment_authorization_rate, successful_payments, unique_payment_attempts\n" +
	"  DURING last_month\n" +
	"  COMPARE TO previous_period"

// QueryYoY documents the year-over-year ShopifyQL companion.
// Tests never execute this against Admin.
const QueryYoY = "" +
	"FROM payment_attempts\n" +
	"  SHOW payment_authorization_rate, successful_payments, unique_payment_attempts\n" +
	"  DURING last_month\n" +
	"  COMPARE TO last_year"

// Kind names a recorded rate move. These are drift inputs, not verdicts.
type Kind string

const (
	// KindPeriod is the move versus the immediately prior equal-length window.
	KindPeriod Kind = "period"
	// KindYoY is the move versus the same window a year earlier.
	KindYoY Kind = "yoy"
)

// Rate is payment_authorization_rate as a percentage (94.2 means 94.2%).
// Comparisons and drift thresholds are in percentage points.
type Rate float64

// PercentagePoints returns the rate in percentage points.
func (r Rate) PercentagePoints() float64 { return float64(r) }

// Period is the reporting window for one ShopifyQL snapshot.
type Period struct {
	Label string `json:"label,omitempty"`
	Start string `json:"start,omitempty"` // YYYY-MM-DD
	End   string `json:"end,omitempty"`   // YYYY-MM-DD
}

// Snapshot is one recorded payment_authorization_rate observation.
type Snapshot struct {
	Metric             string `json:"metric,omitempty"`
	Rate               Rate   `json:"rate_pp"`
	SuccessfulPayments int    `json:"successful_payments,omitempty"`
	UniqueAttempts     int    `json:"unique_payment_attempts,omitempty"`
	Period             Period `json:"period"`
}

// Delta is a period or YoY rate move (current − compare).
// Negative DeltaPP is a drop. A delta is a drift input, not a fraud verdict.
type Delta struct {
	Kind          Kind    `json:"kind"`
	DeltaPP       float64 `json:"delta_pp"`
	Current       Rate    `json:"current_pp"`
	Compare       Rate    `json:"compare_pp"`
	CurrentPeriod Period  `json:"current_period"`
	ComparePeriod Period  `json:"compare_period"`
}

// DropPP is the magnitude of a rate decline in percentage points (0 if no drop).
func (d Delta) DropPP() float64 {
	if d.DeltaPP < 0 {
		return -d.DeltaPP
	}
	return 0
}

// Record is a ShopifyQL auth-rate baseline plus optional period/YoY deltas.
type Record struct {
	Metric  string   `json:"metric"`
	Current Snapshot `json:"current"`
	Period  *Delta   `json:"period,omitempty"`
	YoY     *Delta   `json:"yoy,omitempty"`
}

// Source is a fixture or operator export of ShopifyQL snapshots.
// Load/Parse turn this into a Record with computed period/YoY deltas.
type Source struct {
	Metric      string    `json:"metric,omitempty"`
	Current     Snapshot  `json:"current"`
	PeriodPrior *Snapshot `json:"period_prior,omitempty"`
	YoYPrior    *Snapshot `json:"yoy_prior,omitempty"`
}

// FromCounts returns payment_authorization_rate as a percentage:
// 100 * successful / unique (retries counted once, matching ShopifyQL).
func FromCounts(successful, unique int) (Rate, error) {
	if unique <= 0 {
		return 0, fmt.Errorf("baseline: unique_payment_attempts must be > 0")
	}
	if successful < 0 {
		return 0, fmt.Errorf("baseline: successful_payments must be >= 0")
	}
	if successful > unique {
		return 0, fmt.Errorf("baseline: successful_payments (%d) exceed unique_payment_attempts (%d)", successful, unique)
	}
	return Rate(100 * float64(successful) / float64(unique)), nil
}

// NewDelta records current − compare for a period or YoY move.
func NewDelta(kind Kind, current, compare Snapshot) (Delta, error) {
	if kind != KindPeriod && kind != KindYoY {
		return Delta{}, fmt.Errorf("baseline: unknown delta kind %q", kind)
	}
	if err := current.Valid(); err != nil {
		return Delta{}, err
	}
	if err := compare.Valid(); err != nil {
		return Delta{}, err
	}
	return Delta{
		Kind:          kind,
		DeltaPP:       float64(current.Rate - compare.Rate),
		Current:       current.Rate,
		Compare:       compare.Rate,
		CurrentPeriod: current.Period,
		ComparePeriod: compare.Period,
	}, nil
}

// NewRecord builds a baseline from the current snapshot and optional priors.
func NewRecord(current Snapshot, periodPrior, yoyPrior *Snapshot) (Record, error) {
	if err := current.normalize(""); err != nil {
		return Record{}, err
	}
	rec := Record{Metric: current.Metric, Current: current}
	if periodPrior != nil {
		prior := *periodPrior
		if err := prior.normalize(current.Metric); err != nil {
			return Record{}, fmt.Errorf("baseline: period_prior: %w", err)
		}
		d, err := NewDelta(KindPeriod, current, prior)
		if err != nil {
			return Record{}, err
		}
		rec.Period = &d
	}
	if yoyPrior != nil {
		prior := *yoyPrior
		if err := prior.normalize(current.Metric); err != nil {
			return Record{}, fmt.Errorf("baseline: yoy_prior: %w", err)
		}
		d, err := NewDelta(KindYoY, current, prior)
		if err != nil {
			return Record{}, err
		}
		rec.YoY = &d
	}
	return rec, nil
}

// Parse decodes a Source fixture (or operator export) into a Record.
func Parse(data []byte) (Record, error) {
	var src Source
	if err := json.Unmarshal(data, &src); err != nil {
		return Record{}, fmt.Errorf("baseline: parse: %w", err)
	}
	return src.Record()
}

// Load reads a Source JSON file (typically testdata) and returns a Record.
func Load(path string) (Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, fmt.Errorf("baseline: load %s: %w", path, err)
	}
	return Parse(data)
}

// Record computes period/YoY deltas from the exported snapshots.
func (s Source) Record() (Record, error) {
	metric := strings.TrimSpace(s.Metric)
	if metric == "" {
		metric = Metric
	}
	current := s.Current
	if err := current.normalize(metric); err != nil {
		return Record{}, fmt.Errorf("baseline: current: %w", err)
	}
	return NewRecord(current, s.PeriodPrior, s.YoYPrior)
}

// Valid reports whether the snapshot can be used as a baseline input.
func (s Snapshot) Valid() error {
	dup := s
	return dup.normalize("")
}

func (s *Snapshot) normalize(fallbackMetric string) error {
	if s.Metric == "" {
		s.Metric = fallbackMetric
	}
	if s.Metric == "" {
		s.Metric = Metric
	}
	if s.UniqueAttempts > 0 {
		computed, err := FromCounts(s.SuccessfulPayments, s.UniqueAttempts)
		if err != nil {
			return err
		}
		if s.Rate != 0 && math.Abs(float64(s.Rate-computed)) > 0.05 {
			return fmt.Errorf("baseline: rate_pp %.4f disagrees with counts %.4f", s.Rate, computed)
		}
		s.Rate = computed
	}
	if s.Rate < 0 || s.Rate > 100 {
		return fmt.Errorf("baseline: rate_pp must be in [0, 100], got %v", s.Rate)
	}
	if err := s.Period.valid(); err != nil {
		return err
	}
	return nil
}

func (p Period) valid() error {
	if err := validDate(p.Start); err != nil {
		return fmt.Errorf("baseline: period.start: %w", err)
	}
	if err := validDate(p.End); err != nil {
		return fmt.Errorf("baseline: period.end: %w", err)
	}
	if p.Start != "" && p.End != "" && p.End < p.Start {
		return fmt.Errorf("baseline: period.end %q precedes start %q", p.End, p.Start)
	}
	return nil
}

func validDate(s string) error {
	if s == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("want YYYY-MM-DD, got %q", s)
	}
	return nil
}
