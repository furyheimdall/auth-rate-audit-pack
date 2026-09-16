# How to run the pack

Thin OSS Go core. Tests and `arap pack` use **fixtures only** — no live Admin, no ShopifyQL network, no vendor consoles.

```bash
go test ./...
go run ./cmd/arap pack
go run ./cmd/arap pack -fixture drift
go run ./cmd/arap pack -format pdf -o report.pdf
```

`-fixture` is `healthy` (default) or `drift`. `-format` is `markdown` (default) or `pdf`.

The scored markdown/PDF is the agency SOW attach. It is a checklist score, **not** a fraud verdict.

## Who this is for

Plus / high-mid DTC **not** on Shopify Payments; cards on Adyen/Braintree/Stripe (+ Signifyd/NoFraud/Kount). Buyer: Head of Payments / VP Eng / Plus ops; agency payments lead = influencer. Pre-BFCM / Q4 seasonality only briefly.

## IN (day-1)

1. ShopifyQL auth-rate baseline + period/YoY delta (flag e.g. >0.5pp unexplained drop)
2. Admin webhook / OrderRiskAssessment subscription inventory vs expected
3. Fraud vendor re-verify runbook (Signifyd/NoFraud/Kount — session vs legacy)
4. Staging pass-fail matrix template (Shop Pay/AP/GP + 3DS2 SCA smoke — checklist, not full QA)
5. Scored PDF/report for agency SOW attach

## OUT

full gateway · Tokens/orchestration · full fraud platform · live vendor console automation · gift address-confirm · Deadbugz/DRC/OTM · Shopify Payments-native as primary ICP

## Anchors

**Auth rate drifts.** A ShopifyQL auth-rate period/YoY move (e.g. >0.5pp unexplained drop) is a drift flag, not a fraud verdict.

**Webhooks lie quietly.** Admin webhook / OrderRiskAssessment inventory finds gaps; inventory ≠ remediation.
