package drift

import "testing"

func TestSmoke(t *testing.T) {
	if Seat != "drift" {
		t.Fatalf("Seat = %q", Seat)
	}
	if ThresholdPP != 0.5 {
		t.Fatalf("ThresholdPP = %v", ThresholdPP)
	}
}
