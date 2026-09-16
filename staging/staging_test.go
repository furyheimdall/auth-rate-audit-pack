package staging

import "testing"

func TestSmoke(t *testing.T) {
	if Seat != "staging" {
		t.Fatalf("Seat = %q", Seat)
	}
}
