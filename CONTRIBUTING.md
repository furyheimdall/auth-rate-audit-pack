# Contributing

English README canon is the single source of truth for scope. Keep the locked package seats. Do not open PRs for OUT scope.

## Keep these seats

| Seat | Path |
| --- | --- |
| ShopifyQL `payment_authorization_rate` baseline | `baseline/` |
| Drift flag (>0.5pp) | `drift/` |
| Admin risk / webhook inventory | `inventory/` |
| Vendor re-verify + staging pass/fail | `staging/` |
| Agency / Plus pre-BFCM pack | `pack/` |
| CLI | `cmd/arap/` |

Add implementation inside an existing seat. Do not invent sibling product surfaces.

## Do not open PRs for OUT scope

OUT of the locked MVP (see README):

- Full gateway / Tokens orchestration
- Full fraud platform
- Shopify Payments-native as primary ICP
- Gift address flows
- Deadbugz / DRC / OTM coupling

PRs that add those surfaces will be closed.

## Anchors

- Auth rate drift ≠ fraud verdict
- Inventory ≠ remediation

## Develop

```bash
go test ./...
```

Tests must not require live network.

## Issues

Work is tracked under the `[Epic] Auth Rate Audit Pack MVP` parent and its child issues. Reference the matching child in the PR body. Do not close Epic or child issues from the scaffold PR.
