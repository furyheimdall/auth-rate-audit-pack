package baseline

import "testing"

func TestSmoke(t *testing.T) {
	if Seat != "baseline" {
		t.Fatalf("Seat = %q", Seat)
	}
}
