// Package baseline is the ShopifyQL payment_authorization_rate baseline seat.
//
// Auth rate drift ≠ fraud verdict — this seat records a rate baseline, not a
// fraud decision.
package baseline

// Seat is the locked package seat name.
const Seat = "baseline"
