// Command arap is the Auth Rate Audit Pack CLI.
// It prints seat names, help, or assembles a scored pack from fixtures.
// No network I/O.
package main

import (
	"flag"
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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, w, errw io.Writer) int {
	cmd := "seats"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "help", "-h", "--help":
		return writeHelp(w)
	case "seats", "":
		return writeSeats(w)
	case "pack":
		return runPack(args[1:], w, errw)
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
  arap pack     assemble scored report from fixtures (markdown or pdf)
  arap help     print this help

Pack flags:
  -fixture name    healthy (default) or drift
  -format name     markdown (default) or pdf
  -o path          write to path (default stdout)

Fixtures only — no network. Thin OSS Go core — not a full fraud platform, not a gateway / Tokens orchestration.
See docs/run.md.
`)
	return 0
}

func runPack(args []string, w, errw io.Writer) int {
	fs := flag.NewFlagSet("pack", flag.ContinueOnError)
	fs.SetOutput(errw)
	fixture := fs.String("fixture", "healthy", "fixture name (healthy|drift)")
	format := fs.String("format", "markdown", "report format (markdown|pdf)")
	out := fs.String("o", "", "output path (default stdout)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	fx, err := pack.LoadFixture(pack.FixtureFS(), *fixture)
	if err != nil {
		fmt.Fprintln(errw, err)
		return 2
	}
	report, err := pack.Assemble(fx.Sources())
	if err != nil {
		fmt.Fprintln(errw, err)
		return 1
	}

	dest := w
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		defer f.Close()
		dest = f
	}
	if err := report.Write(dest, pack.Format(*format)); err != nil {
		fmt.Fprintln(errw, err)
		return 1
	}
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
