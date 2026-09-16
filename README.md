# Auth Rate Audit Pack

Q4 auth-rate baseline + fraud/order-webhook mismatch checklist for Shopify Plus on Adyen/Braintree/Stripe — not a gateway.

**Auth rate drifts. Webhooks lie quietly.**

English is the single source of truth for this marketing MVP canon. Do not invent scope beyond the locked IN / OUT tables.

This repository is a thin OSS Go core. It is **not a gateway** and **not a full fraud platform**.

## Anchors

**Auth rate drifts. Webhooks lie quietly.**

Secondary (technical, not hero copy):

| Note | Meaning |
| --- | --- |
| Auth rate drift ≠ fraud verdict | A >0.5pp unexplained drop vs baseline is a flag, not a fraud decision. |
| Inventory ≠ remediation | Listing webhook / OrderRiskAssessment gaps is not fixing them. |

## ICP

Plus / high-mid DTC not on Shopify Payments; cards on Adyen/Braintree/Stripe (+ Signifyd/NoFraud/Kount). Buyer: Head of Payments / VP Eng / Plus ops; agency payments lead = influencer.

## IN (day-1)

| # | In scope |
| --- | --- |
| 1 | ShopifyQL auth-rate baseline + period/YoY delta (flag e.g. >0.5pp unexplained drop) |
| 2 | Admin webhook / OrderRiskAssessment subscription inventory vs expected |
| 3 | Fraud vendor re-verify runbook (Signifyd/NoFraud/Kount — session vs legacy) |
| 4 | Staging pass-fail matrix template (Shop Pay/AP/GP + 3DS2 SCA smoke — checklist, not full QA) |
| 5 | Scored PDF/report for agency SOW attach |

## OUT

| Out of scope | Note |
| --- | --- |
| Full gateway | Out — reject PRs |
| Tokens / orchestration | Out — reject PRs |
| Full fraud platform | Out — reject PRs |
| Live vendor console automation | Out — reject PRs |
| Gift address-confirm | Out — reject PRs |
| Deadbugz / DRC / OTM coupling | Out — reject PRs |

## Package seats

| Package | Seat |
| --- | --- |
| [`baseline/`](baseline/) | ShopifyQL auth-rate baseline + period/YoY delta |
| [`drift/`](drift/) | Flag e.g. >0.5pp unexplained drop |
| [`inventory/`](inventory/) | Admin webhook / OrderRiskAssessment subscription inventory vs expected |
| [`staging/`](staging/) | Fraud vendor re-verify runbook + staging pass-fail matrix |
| [`pack/`](pack/) | Scored PDF/report for agency SOW attach |
| [`cmd/arap/`](cmd/arap/) | CLI stub (prints seat names or help; no network in tests) |

Day-1 seats are stubs that compile and pass smoke tests. Tests must not use the network. Do not add OUT-scope packages.

## Develop

```bash
go test ./...
```

CLI (no network):

```bash
go run ./cmd/arap
go run ./cmd/arap help
go run ./cmd/arap seats
```

Module: [`github.com/furyheimdall/auth-rate-audit-pack`](https://github.com/furyheimdall/auth-rate-audit-pack) · License: [MIT](LICENSE)

See [CONTRIBUTING.md](CONTRIBUTING.md).
