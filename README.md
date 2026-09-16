# Auth Rate Audit Pack

Q4 auth-rate baseline + fraud/order-webhook mismatch checklist for Shopify Plus on Adyen/Braintree/Stripe — not a gateway.

**Auth rate drifts. Webhooks lie quietly.**

English is the single source of truth. Thin OSS Go core — not a full fraud platform, not a gateway / Tokens orchestration.

## Anchors

| Anchor | Meaning |
| --- | --- |
| Auth rate drifts. | ShopifyQL auth-rate period/YoY move (e.g. >0.5pp unexplained drop) is a drift flag, not a fraud verdict. |
| Webhooks lie quietly. | Admin webhook / OrderRiskAssessment inventory finds gaps; inventory ≠ remediation. |

## ICP

Plus / high-mid DTC **not** on Shopify Payments; cards on Adyen/Braintree/Stripe (+ Signifyd/NoFraud/Kount). Buyer: Head of Payments / VP Eng / Plus ops; agency payments lead = influencer. Pre-BFCM / Q4 seasonality only briefly.

## IN (day-1)

1. ShopifyQL auth-rate baseline + period/YoY delta (flag e.g. >0.5pp unexplained drop)
2. Admin webhook / OrderRiskAssessment subscription inventory vs expected
3. Fraud vendor re-verify runbook (Signifyd/NoFraud/Kount — session vs legacy)
4. Staging pass-fail matrix template (Shop Pay/AP/GP + 3DS2 SCA smoke — checklist, not full QA)
5. Scored PDF/report for agency SOW attach

## OUT

full gateway · Tokens/orchestration · full fraud platform · live vendor console automation · gift address-confirm · Deadbugz/DRC/OTM · Shopify Payments-native as primary ICP

## Package seats

| Package | Seat |
| --- | --- |
| [`baseline/`](baseline/) | ShopifyQL auth-rate baseline + period/YoY delta |
| [`drift/`](drift/) | Flag e.g. >0.5pp unexplained drop |
| [`inventory/`](inventory/) | Admin webhook / OrderRiskAssessment subscription inventory vs expected |
| [`staging/`](staging/) | Fraud vendor re-verify runbook (Signifyd/NoFraud/Kount — session vs legacy) + staging pass-fail matrix (Shop Pay/AP/GP + 3DS2 SCA smoke) |
| [`pack/`](pack/) | Scored PDF/report for agency SOW attach — assembles the other seats from interfaces/fixtures |
| [`cmd/arap/`](cmd/arap/) | CLI (seats, help, `pack`; no network in tests) |

`baseline/` and `drift/` are fixture-backed (no live Admin/ShopifyQL). `inventory/` and `staging/` may still be stubs. `pack/` assembles the seats from interfaces + fixtures so the scored report can ship without live Admin/ShopifyQL. Tests must not use the network. Do not add OUT-scope packages.

## Develop

```bash
go test ./...
```

CLI (no network; pack uses fixtures):

```bash
go run ./cmd/arap
go run ./cmd/arap help
go run ./cmd/arap seats
go run ./cmd/arap pack
go run ./cmd/arap pack -fixture drift -format pdf -o report.pdf
```

How to run the pack, plus IN/OUT reminders: [docs/run.md](docs/run.md).

Module: [`github.com/furyheimdall/auth-rate-audit-pack`](https://github.com/furyheimdall/auth-rate-audit-pack) · License: [MIT](LICENSE)

See [CONTRIBUTING.md](CONTRIBUTING.md).
