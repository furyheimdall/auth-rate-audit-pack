package inventory

import "testing"

func TestSmoke(t *testing.T) {
	if Seat != "inventory" {
		t.Fatalf("Seat = %q", Seat)
	}
}
