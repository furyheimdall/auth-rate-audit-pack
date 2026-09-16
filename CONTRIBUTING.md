# Contributing

English README canon is the single source of truth for scope. Keep the locked package seats. Do not open PRs for OUT scope.

## Keep these seats

| Seat | Path |
| --- | --- |
| ShopifyQL auth-rate baseline + period/YoY delta | `baseline/` |
| Flag e.g. >0.5pp unexplained drop | `drift/` |
| Admin webhook / OrderRiskAssessment inventory vs expected | `inventory/` |
| Fraud vendor re-verify runbook + staging pass-fail matrix | `staging/` |
| Scored PDF/report for agency SOW attach | `pack/` |
| CLI | `cmd/arap/` |

Add implementation inside an existing seat. Do not invent sibling product surfaces.

## Do not open PRs for OUT scope

OUT of the locked MVP (see README):

- full gateway
- Tokens/orchestration
- full fraud platform
- live vendor console automation
- gift address-confirm
- Deadbugz/DRC/OTM
- Shopify Payments-native as primary ICP

PRs that add those surfaces will be closed.

## Anchors

**Auth rate drifts. Webhooks lie quietly.**

| Anchor | Meaning |
| --- | --- |
| Auth rate drifts. | ShopifyQL auth-rate period/YoY move (e.g. >0.5pp unexplained drop) is a drift flag, not a fraud verdict. |
| Webhooks lie quietly. | Admin webhook / OrderRiskAssessment inventory finds gaps; inventory ≠ remediation. |

## Develop

```bash
go test ./...
go run ./cmd/arap pack
```

Tests must not require live network. Pack assembly uses fixtures when other seats are still stubs. Do not automate live vendor consoles. See [docs/run.md](docs/run.md).

## Issues

Work is tracked under the `[Epic] Auth Rate Audit Pack MVP` parent and its child issues. Reference the matching child in the PR body. Do not close Epic or child issues from the scaffold PR.
