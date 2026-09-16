package main

import (
	"strings"
	"testing"
)

func TestSeats(t *testing.T) {
	var b strings.Builder
	if code := run(nil, &b); code != 0 {
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
	if code := run([]string{"help"}, &b); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(b.String(), "Auth rate drift ≠ fraud verdict") {
		t.Fatalf("help = %q", b.String())
	}
}

func TestUnknown(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"fraud"}, &b); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestSeatsList(t *testing.T) {
	if len(seats()) != 5 {
		t.Fatalf("seats = %v", seats())
	}
}
