// Package inventory is the Admin risk / webhook inventory checklist seat.
//
// Webhooks lie quietly. Inventory finds gaps; inventory ≠ remediation.
package inventory

// Seat is the locked package seat name.
const Seat = "inventory"

// Kind is an inventory item type.
type Kind string

const (
	// KindWebhook is an Admin webhook subscription (topic).
	KindWebhook Kind = "admin_webhook"
	// KindOrderRiskAssessment is an OrderRiskAssessment provider subscription.
	KindOrderRiskAssessment Kind = "order_risk_assessment"
)

// Locked Admin webhook topics inventoried against the expected set.
const (
	TopicOrdersCreate                = "ORDERS_CREATE"
	TopicOrdersUpdated               = "ORDERS_UPDATED"
	TopicOrdersPaid                  = "ORDERS_PAID"
	TopicOrdersCancelled             = "ORDERS_CANCELLED"
	TopicOrdersRiskAssessmentChanged = "ORDERS_RISK_ASSESSMENT_CHANGED"
)

// Status is one expected-vs-actual checklist outcome.
type Status string

const (
	// StatusMatch is expected and present.
	StatusMatch Status = "match"
	// StatusGap is expected but missing. A gap is a finding, not a fix.
	StatusGap Status = "gap"
	// StatusUnexpected is present but not in the expected set.
	StatusUnexpected Status = "unexpected"
)

// Item is one expected or observed subscription.
type Item struct {
	Kind Kind   `json:"kind"`
	Key  string `json:"key"`
}

// Row is one expected-vs-actual checklist line.
type Row struct {
	Kind     Kind   `json:"kind"`
	Key      string `json:"key"`
	Expected bool   `json:"expected"`
	Actual   bool   `json:"actual"`
	Status   Status `json:"status"`
}

// Checklist is inventory output only — listing a gap is not remediating it.
type Checklist struct {
	Rows []Row `json:"rows"`
}

type itemKey struct {
	Kind Kind
	Key  string
}

// DefaultExpected is the locked Admin webhook set for this pack.
// OrderRiskAssessment provider keys are shop-specific; add them on the
// expected fixture rather than assuming every vendor is subscribed.
func DefaultExpected() []Item {
	return []Item{
		{Kind: KindWebhook, Key: TopicOrdersCreate},
		{Kind: KindWebhook, Key: TopicOrdersUpdated},
		{Kind: KindWebhook, Key: TopicOrdersPaid},
		{Kind: KindWebhook, Key: TopicOrdersCancelled},
		{Kind: KindWebhook, Key: TopicOrdersRiskAssessmentChanged},
	}
}

// Check builds an expected-vs-actual checklist. No live Admin API.
func Check(expected, actual []Item) Checklist {
	want := indexItems(expected)
	have := indexItems(actual)
	seen := make(map[itemKey]bool, len(want)+len(have))
	rows := make([]Row, 0, len(want)+len(have))

	for _, it := range expected {
		k := itemKey{it.Kind, it.Key}
		if seen[k] {
			continue
		}
		seen[k] = true
		_, ok := have[k]
		st := StatusMatch
		if !ok {
			st = StatusGap
		}
		rows = append(rows, Row{
			Kind:     it.Kind,
			Key:      it.Key,
			Expected: true,
			Actual:   ok,
			Status:   st,
		})
	}
	for _, it := range actual {
		k := itemKey{it.Kind, it.Key}
		if seen[k] {
			continue
		}
		seen[k] = true
		rows = append(rows, Row{
			Kind:     it.Kind,
			Key:      it.Key,
			Expected: false,
			Actual:   true,
			Status:   StatusUnexpected,
		})
	}
	return Checklist{Rows: rows}
}

// Gaps returns expected-but-missing rows. Inventory ≠ remediation.
func (c Checklist) Gaps() []Row {
	return c.filter(StatusGap)
}

// Unexpected returns actual-but-not-expected rows.
func (c Checklist) Unexpected() []Row {
	return c.filter(StatusUnexpected)
}

func (c Checklist) filter(st Status) []Row {
	out := make([]Row, 0)
	for _, r := range c.Rows {
		if r.Status == st {
			out = append(out, r)
		}
	}
	return out
}

func indexItems(items []Item) map[itemKey]Item {
	m := make(map[itemKey]Item, len(items))
	for _, it := range items {
		m[itemKey{it.Kind, it.Key}] = it
	}
	return m
}
