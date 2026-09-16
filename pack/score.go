package pack

import (
	"fmt"
	"math"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
	"github.com/furyheimdall/auth-rate-audit-pack/drift"
	"github.com/furyheimdall/auth-rate-audit-pack/inventory"
	"github.com/furyheimdall/auth-rate-audit-pack/staging"
)

// Equal stub weights — agency SOW attach score, not a fraud model.
const (
	WeightBaseline  = 25
	WeightDrift     = 25
	WeightInventory = 25
	WeightStaging   = 25
)

// Audience is who the scored pack is written for.
const Audience = "Head of Payments / VP Eng / Plus ops; agency payments lead = influencer."

// ICP is the locked day-1 customer. Shopify Payments-native is not primary.
const ICP = "Plus / high-mid DTC not on Shopify Payments; cards on Adyen/Braintree/Stripe (+ Signifyd/NoFraud/Kount)."

// SeasonNote is the brief pre-BFCM / Q4 reminder — not a capacity forecast.
const SeasonNote = "Pre-BFCM / Q4 seasonality is brief context only; this pack is not a peak-capacity forecast."

// AnchorAuthRate is the locked auth-rate anchor.
const AnchorAuthRate = "Auth rate drifts. A ShopifyQL auth-rate period/YoY move (e.g. >0.5pp unexplained drop) is a drift flag, not a fraud verdict."

// AnchorWebhooks is the locked webhook-inventory anchor.
const AnchorWebhooks = "Webhooks lie quietly. Admin webhook / OrderRiskAssessment inventory finds gaps; inventory ≠ remediation."

// OUT is the locked MVP exclusion line.
const OUT = "full gateway · Tokens/orchestration · full fraud platform · live vendor console automation · gift address-confirm · Deadbugz/DRC/OTM · Shopify Payments-native as primary ICP"

// IN lists the locked day-1 surfaces this pack assembles.
var IN = []string{
	"ShopifyQL auth-rate baseline + period/YoY delta (flag e.g. >0.5pp unexplained drop)",
	"Admin webhook / OrderRiskAssessment subscription inventory vs expected",
	"Fraud vendor re-verify runbook (Signifyd/NoFraud/Kount — session vs legacy)",
	"Staging pass-fail matrix template (Shop Pay/AP/GP + 3DS2 SCA smoke — checklist, not full QA)",
	"Scored PDF/report for agency SOW attach",
}

// Section is one seat's contribution to the pack score.
type Section struct {
	Seat   string
	Score  int
	Weight int
	Notes  string
}

// Finding is a human-readable pack note. Flags and gaps are not remediations.
type Finding struct {
	Seat     string
	Severity string // info | flag | gap
	Text     string
}

// Report is the scored pack for agency SOW attach.
type Report struct {
	Title      string
	Audience   string
	ICP        string
	SeasonNote string
	Overall    int
	Sections   []Section
	Findings   []Finding
	Inputs     Inputs
	IN         []string
	OUT        string
}

// Score builds a deterministic stub score from already-collected inputs.
func Score(in Inputs) Report {
	sections := []Section{
		scoreBaseline(in.Baseline),
		scoreDrift(in.Drift),
		scoreInventory(in.Inventory),
		scoreStaging(in.Staging),
	}
	return Report{
		Title:      "Auth Rate Audit Pack — scored report",
		Audience:   Audience,
		ICP:        ICP,
		SeasonNote: SeasonNote,
		Overall:    weightedAverage(sections),
		Sections:   sections,
		Findings:   findings(in),
		Inputs:     in,
		IN:         append([]string(nil), IN...),
		OUT:        OUT,
	}
}

func scoreBaseline(b BaselineSnapshot) Section {
	s := Section{
		Seat:   baseline.Seat,
		Weight: WeightBaseline,
		Notes:  "ShopifyQL payment_authorization_rate baseline + period/YoY delta. Fixture-backed until a live query lands. Not a fraud verdict.",
	}
	if !b.Recorded() {
		s.Score = 0
		s.Notes += " Missing snapshot."
		return s
	}
	s.Score = 100
	return s
}

func scoreDrift(d DriftFlag) Section {
	s := Section{
		Seat:   drift.Seat,
		Weight: WeightDrift,
		Notes:  fmt.Sprintf("Flag e.g. >%.1fpp unexplained drop. Auth rate drift ≠ fraud verdict.", d.ThresholdPP),
	}
	if !d.Recorded() {
		s.Score = 0
		s.Notes += " Missing flag."
		return s
	}
	switch {
	case !d.Flagged:
		s.Score = 100
	case d.Unexplained:
		s.Score = 40
	default:
		s.Score = 70
	}
	return s
}

func scoreInventory(c InventoryChecklist) Section {
	s := Section{
		Seat:   inventory.Seat,
		Weight: WeightInventory,
		Notes:  "Admin webhook / OrderRiskAssessment inventory vs expected. Inventory ≠ remediation.",
	}
	if len(c.Items) == 0 {
		s.Score = 0
		s.Notes += " No checklist rows."
		return s
	}
	expected := 0
	present := 0
	for _, it := range c.Items {
		if !it.Expected {
			continue
		}
		expected++
		if it.Present {
			present++
		}
	}
	if expected == 0 {
		s.Score = 0
		return s
	}
	s.Score = int(math.Round(100 * float64(present) / float64(expected)))
	return s
}

func scoreStaging(m StagingMatrix) Section {
	s := Section{
		Seat:   staging.Seat,
		Weight: WeightStaging,
		Notes:  "Vendor re-verify (Signifyd/NoFraud/Kount — session vs legacy) + Shop Pay/AP/GP + 3DS2 SCA smoke. Checklist, not full QA.",
	}
	pass, fail := 0, 0
	for _, v := range m.Vendors {
		switch v.Status {
		case StatusPass:
			pass++
		case StatusFail:
			fail++
		}
	}
	for _, c := range m.Checks {
		switch c.Status {
		case StatusPass:
			pass++
		case StatusFail:
			fail++
		}
	}
	total := pass + fail
	if total == 0 {
		s.Score = 0
		s.Notes += " No scored rows (pass/fail)."
		return s
	}
	s.Score = int(math.Round(100 * float64(pass) / float64(total)))
	return s
}

func weightedAverage(sections []Section) int {
	total, weight := 0, 0
	for _, s := range sections {
		total += s.Score * s.Weight
		weight += s.Weight
	}
	if weight == 0 {
		return 0
	}
	return (total + weight/2) / weight
}

func findings(in Inputs) []Finding {
	var out []Finding
	out = append(out, Finding{
		Seat:     baseline.Seat,
		Severity: "info",
		Text:     AnchorAuthRate,
	})
	out = append(out, Finding{
		Seat:     inventory.Seat,
		Severity: "info",
		Text:     AnchorWebhooks,
	})
	if gw := in.Baseline.Gateway; gw == "Shopify Payments" {
		out = append(out, Finding{
			Seat:     baseline.Seat,
			Severity: "flag",
			Text:     "Gateway is Shopify Payments; primary ICP is Adyen/Braintree/Stripe — Shopify Payments-native is out as primary ICP.",
		})
	}
	if in.Drift.Flagged {
		out = append(out, Finding{
			Seat:     drift.Seat,
			Severity: "flag",
			Text: fmt.Sprintf(
				"Drift flagged: observed %.2fpp vs threshold %.1fpp (unexplained=%t). This is a drift flag, not a fraud verdict.",
				in.Drift.ObservedDeltaPP, in.Drift.ThresholdPP, in.Drift.Unexplained,
			),
		})
	}
	for _, g := range in.Inventory.Gaps() {
		out = append(out, Finding{
			Seat:     inventory.Seat,
			Severity: "gap",
			Text:     fmt.Sprintf("Inventory gap: %s expected but not present. Listing a gap is not remediating it.", g.Topic),
		})
	}
	for _, v := range in.Staging.Vendors {
		if v.Status == StatusFail {
			out = append(out, Finding{
				Seat:     staging.Seat,
				Severity: "flag",
				Text:     fmt.Sprintf("Vendor re-verify %s (%s) failed. Runbook/checklist only — no live vendor console automation.", v.Vendor, v.Mode),
			})
		}
	}
	for _, c := range in.Staging.Checks {
		if c.Status == StatusFail {
			out = append(out, Finding{
				Seat:     staging.Seat,
				Severity: "flag",
				Text:     fmt.Sprintf("Staging check %s failed. Smoke checklist, not full QA.", c.Name),
			})
		}
	}
	return out
}
