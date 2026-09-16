// Package drift flags authorization-rate movement versus a baseline.
//
// Auth rate drift ≠ fraud verdict — a >0.5pp flag is not a fraud decision.
package drift

// Seat is the locked package seat name.
const Seat = "drift"

// ThresholdPP is the locked day-1 drift flag: rate moves >0.5 percentage
// points versus baseline.
const ThresholdPP = 0.5
