// Command arap is the Auth Rate Audit Pack CLI stub.
// It prints seat names or help. No network I/O.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/furyheimdall/auth-rate-audit-pack/baseline"
	"github.com/furyheimdall/auth-rate-audit-pack/drift"
	"github.com/furyheimdall/auth-rate-audit-pack/inventory"
	"github.com/furyheimdall/auth-rate-audit-pack/pack"
	"github.com/furyheimdall/auth-rate-audit-pack/staging"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, w io.Writer) int {
	cmd := "seats"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "help", "-h", "--help":
		return writeHelp(w)
	case "seats", "":
		return writeSeats(w)
	default:
		fmt.Fprintf(w, "unknown command %q\n", cmd)
		writeHelp(w)
		return 2
	}
}

func writeHelp(w io.Writer) int {
	fmt.Fprint(w, `Auth Rate Audit Pack — arap

Q4 auth-rate baseline + fraud/order-webhook mismatch checklist for Shopify Plus on Adyen/Braintree/Stripe — not a gateway.
Auth rate drifts. Webhooks lie quietly.

Usage:
  arap          print package seats
  arap seats    print package seats
  arap help     print this help

No network. Thin OSS Go core — not a full fraud platform, not a gateway / Tokens orchestration.
`)
	return 0
}

func writeSeats(w io.Writer) int {
	for _, name := range seats() {
		fmt.Fprintln(w, name)
	}
	return 0
}

func seats() []string {
	return []string{
		baseline.Seat,
		drift.Seat,
		inventory.Seat,
		staging.Seat,
		pack.Seat,
	}
}
