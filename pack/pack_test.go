package pack

import "testing"

func TestSmoke(t *testing.T) {
	if Seat != "pack" {
		t.Fatalf("Seat = %q", Seat)
	}
}
