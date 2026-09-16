package pack

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Format is a pack render format.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatPDF      Format = "pdf"
)

// Write renders the scored report. markdown is the agency-readable attach;
// pdf is a one-page stub of the same score (not a designed layout product).
func (r Report) Write(w io.Writer, format Format) error {
	switch format {
	case FormatMarkdown, "":
		_, err := io.WriteString(w, r.Markdown())
		return err
	case FormatPDF:
		return r.WritePDF(w)
	default:
		return fmt.Errorf("pack: unknown format %q (markdown|pdf)", format)
	}
}

// Markdown is the thin scored report for agency SOW attach.
func (r Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", r.Title)
	fmt.Fprintf(&b, "**Overall score:** %d / 100\n\n", r.Overall)
	fmt.Fprintf(&b, "**Audience:** %s\n\n", r.Audience)
	fmt.Fprintf(&b, "**ICP:** %s\n\n", r.ICP)
	fmt.Fprintf(&b, "**Seasonality:** %s\n\n", r.SeasonNote)
	b.WriteString("This scored checklist is for agency SOW attach. It is not a fraud verdict, not a gateway, and not a full fraud platform.\n\n")

	b.WriteString("## Anchors\n\n")
	fmt.Fprintf(&b, "- %s\n", AnchorAuthRate)
	fmt.Fprintf(&b, "- %s\n\n", AnchorWebhooks)

	b.WriteString("## Section scores\n\n")
	b.WriteString("| Seat | Score | Weight | Notes |\n| --- | --- | --- | --- |\n")
	for _, s := range r.Sections {
		fmt.Fprintf(&b, "| %s | %d | %d | %s |\n", s.Seat, s.Score, s.Weight, s.Notes)
	}
	b.WriteString("\n")

	in := r.Inputs
	b.WriteString("## Baseline\n\n")
	fmt.Fprintf(&b, "- Shop: %s\n", dash(in.Baseline.Shop))
	fmt.Fprintf(&b, "- Gateway: %s\n", dash(in.Baseline.Gateway))
	fmt.Fprintf(&b, "- Period: %s\n", dash(in.Baseline.PeriodLabel))
	fmt.Fprintf(&b, "- Period rate: %.2f%% (prior %.2f%%, delta %+.2fpp)\n", in.Baseline.PeriodRate, in.Baseline.PriorPeriodRate, in.Baseline.PeriodDeltaPP)
	fmt.Fprintf(&b, "- YoY rate: %.2f%% (delta %+.2fpp)\n\n", in.Baseline.YoYRate, in.Baseline.YoYDeltaPP)

	b.WriteString("## Drift\n\n")
	fmt.Fprintf(&b, "- Observed delta: %+.2fpp\n", in.Drift.ObservedDeltaPP)
	fmt.Fprintf(&b, "- Threshold: >%.1fpp unexplained drop\n", in.Drift.ThresholdPP)
	fmt.Fprintf(&b, "- Unexplained: %t\n", in.Drift.Unexplained)
	fmt.Fprintf(&b, "- Flagged: %t — drift flag, not a fraud verdict\n\n", in.Drift.Flagged)

	b.WriteString("## Inventory\n\n")
	if len(in.Inventory.Items) == 0 {
		b.WriteString("No checklist rows.\n\n")
	} else {
		b.WriteString("| Topic | Expected | Present | Gap |\n| --- | --- | --- | --- |\n")
		for _, it := range in.Inventory.Items {
			fmt.Fprintf(&b, "| %s | %t | %t | %t |\n", it.Topic, it.Expected, it.Present, it.Gap())
		}
		b.WriteString("\nInventory finds gaps; inventory ≠ remediation.\n\n")
	}

	b.WriteString("## Staging\n\n")
	b.WriteString("Vendor re-verify (session vs legacy):\n\n")
	if len(in.Staging.Vendors) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, v := range in.Staging.Vendors {
			fmt.Fprintf(&b, "- %s (%s): %s\n", v.Vendor, v.Mode, v.Status)
		}
	}
	b.WriteString("\nPass/fail matrix (Shop Pay/AP/GP + 3DS2 SCA smoke — checklist, not full QA):\n\n")
	if len(in.Staging.Checks) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, c := range in.Staging.Checks {
			fmt.Fprintf(&b, "- %s: %s\n", c.Name, c.Status)
		}
	}
	b.WriteString("\n")

	b.WriteString("## Findings\n\n")
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "- [%s / %s] %s\n", f.Severity, f.Seat, f.Text)
	}
	b.WriteString("\n")

	b.WriteString("## IN (day-1)\n\n")
	for i, line := range r.IN {
		fmt.Fprintf(&b, "%d. %s\n", i+1, line)
	}
	b.WriteString("\n## OUT\n\n")
	fmt.Fprintf(&b, "%s\n", r.OUT)
	return b.String()
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// WritePDF writes a one-page stub PDF of the score. No layout library;
// enough to attach a scored PDF to an agency SOW.
func (r Report) WritePDF(w io.Writer) error {
	lines := r.pdfLines()
	var content strings.Builder
	content.WriteString("BT\n/F1 10 Tf\n48 746 Td\n")
	for i, line := range lines {
		if i > 0 {
			content.WriteString("0 -13 Td\n")
		}
		fmt.Fprintf(&content, "(%s) Tj\n", pdfEscape(line))
	}
	content.WriteString("ET\n")
	stream := content.String()

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offs := make([]int, 6)
	writeObj := func(n int, body string) {
		offs[n] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", n, body)
	}
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")
	writeObj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream))
	writeObj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 6\n0000000000 65535 f \n")
	for n := 1; n <= 5; n++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offs[n])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	_, err := w.Write(buf.Bytes())
	return err
}

func (r Report) pdfLines() []string {
	lines := []string{
		r.Title,
		fmt.Sprintf("Overall score: %d / 100", r.Overall),
		"Agency SOW attach. Not a fraud verdict, not a gateway, not a fraud platform.",
		"Audience: " + ascii(r.Audience),
		"ICP: Plus / high-mid DTC not on Shopify Payments; Adyen/Braintree/Stripe.",
		ascii(r.SeasonNote),
		"Anchors: Auth rate drifts. Webhooks lie quietly.",
		"Section scores:",
	}
	for _, s := range r.Sections {
		lines = append(lines, fmt.Sprintf("  %s  %d/100  weight %d", s.Seat, s.Score, s.Weight))
	}
	in := r.Inputs
	lines = append(lines,
		fmt.Sprintf("Baseline: %s %s %s rate %.2f%% delta %+.2fpp",
			ascii(in.Baseline.Shop), ascii(in.Baseline.Gateway), ascii(in.Baseline.PeriodLabel),
			in.Baseline.PeriodRate, in.Baseline.PeriodDeltaPP),
		fmt.Sprintf("Drift: flagged=%t unexplained=%t observed %+.2fpp threshold %.1fpp",
			in.Drift.Flagged, in.Drift.Unexplained, in.Drift.ObservedDeltaPP, in.Drift.ThresholdPP),
		fmt.Sprintf("Inventory gaps: %d (inventory != remediation)", len(in.Inventory.Gaps())),
		fmt.Sprintf("Staging vendors=%d checks=%d (checklist, not full QA)", len(in.Staging.Vendors), len(in.Staging.Checks)),
		"OUT: no full gateway, Tokens, fraud platform, live vendor consoles,",
		"gift address-confirm, Deadbugz/DRC/OTM, Shopify Payments-native ICP.",
	)
	const max = 48
	if len(lines) > max {
		lines = lines[:max]
	}
	return lines
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return ascii(s)
}

func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 32 || r > 126 {
			b.WriteByte('?')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
