# Contributing

English README canon is the single source of truth for scope. Keep the locked package seats. Do not open PRs for OUT scope.

## Keep these seats

| Seat | Path |
| --- | --- |
| ShopifyQL auth-rate baseline + period/YoY delta | `baseline/` |
| Flag e.g. >0.5pp unexplained drop | `drift/` |
| Admin webhook / OrderRiskAssessment inventory | `inventory/` |
| Fraud vendor re-verify + staging pass-fail matrix | `staging/` |
| Scored PDF/report for agency SOW attach | `pack/` |
| CLI | `cmd/arap/` |

Add implementation inside an existing seat. Do not invent sibling product surfaces.

## Do not open PRs for OUT scope

OUT of the locked MVP (see README):

- Full gateway
- Tokens / orchestration
- Full fraud platform
- Live vendor console automation
- Gift address-confirm
- Deadbugz / DRC / OTM coupling

PRs that add those surfaces will be closed.

## Anchors

**Auth rate drifts. Webhooks lie quietly.**

Secondary (technical): auth rate drift ≠ fraud verdict; inventory ≠ remediation.

## Develop

```bash
go test ./...
```

Tests must not require live network. Do not automate live vendor consoles.

## Issues

Work is tracked under the `[Epic] Auth Rate Audit Pack MVP` parent and its child issues. Reference the matching child in the PR body. Do not close Epic or child issues from the scaffold PR.
