package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeats(t *testing.T) {
	var b strings.Builder
	if code := run(nil, &b, &b); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := b.String()
	for _, seat := range []string{"baseline", "drift", "inventory", "staging", "pack"} {
		if !strings.Contains(got, seat) {
			t.Fatalf("missing seat %q in %q", seat, got)
		}
	}
}

func TestHelp(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"help"}, &b, &b); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := b.String()
	if !strings.Contains(got, "Auth rate drifts. Webhooks lie quietly.") {
		t.Fatalf("help = %q", got)
	}
	if !strings.Contains(got, "arap pack") {
		t.Fatalf("help missing pack: %q", got)
	}
}

func TestUnknown(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"fraud"}, &b, &b); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestPackMarkdown(t *testing.T) {
	var out, errb strings.Builder
	if code := run([]string{"pack", "-fixture", "healthy"}, &out, &errb); code != 0 {
		t.Fatalf("exit %d err=%q", code, errb.String())
	}
	got := out.String()
	if !strings.Contains(got, "**Overall score:** 100 / 100") {
		t.Fatalf("pack = %q", got)
	}
	if !strings.Contains(got, "plus-dtc-adyen-fixture") {
		t.Fatalf("pack missing fixture shop: %q", got)
	}
}

func TestPackPDFFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.pdf")
	var out, errb strings.Builder
	code := run([]string{"pack", "-fixture", "drift", "-format", "pdf", "-o", path}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d err=%q", code, errb.String())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("%PDF-1.4")) {
		t.Fatalf("pdf prefix = %q", raw[:min(16, len(raw))])
	}
}

func TestPackUnknownFixture(t *testing.T) {
	var out, errb strings.Builder
	if code := run([]string{"pack", "-fixture", "live-admin"}, &out, &errb); code != 2 {
		t.Fatalf("exit %d, want 2 err=%q", code, errb.String())
	}
}

func TestSeatsList(t *testing.T) {
	if len(seats()) != 5 {
		t.Fatalf("seats = %v", seats())
	}
}
