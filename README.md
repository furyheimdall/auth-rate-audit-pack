# Auth Rate Audit Pack

ShopifyQL payment_authorization_rate baseline + drift flag + Admin risk/webhook inventory + staging pass/fail — pre-BFCM agency/Plus pack.

**Auth rate drift ≠ fraud verdict. / Inventory ≠ remediation.**

English is the single source of truth for this marketing MVP canon. Do not invent scope beyond the locked IN / OUT tables.

This repository is a thin OSS Go core. It is **not a full fraud platform** and **not a full gateway / Tokens orchestration**.

## Anchors

| Anchor | Meaning |
| --- | --- |
| Auth rate drift ≠ fraud verdict | A >0.5pp move vs baseline is a drift flag, not a fraud decision. |
| Inventory ≠ remediation | Listing Admin risk / webhook gaps is not fixing them. |

## ICP

Agency + Shopify Plus mid-market (pre-BFCM), NOT Shopify Payments-native primary.

## IN (day-1)

| In scope | Locked detail |
| --- | --- |
| ShopifyQL `payment_authorization_rate` baseline | Record a baseline authorization rate |
| Drift flag | Flag when rate moves >0.5 percentage points vs baseline |
| Admin risk / webhook inventory checklist | Inventory only |
| Vendor re-verify + staging pass/fail | Staging result, not production cutover |
| Agency / Plus pre-BFCM pack framing | One pack assembly seat |

## OUT

| Out of scope | Note |
| --- | --- |
| Full gateway / Tokens orchestration | Out — reject PRs |
| Full fraud platform | Out — reject PRs |
| Shopify Payments-native as primary ICP | Out — reject PRs |
| Gift address flows | Out — reject PRs |
| Deadbugz / DRC / OTM coupling | Out — reject PRs |

## Package seats

| Package | Seat |
| --- | --- |
| [`baseline/`](baseline/) | ShopifyQL `payment_authorization_rate` baseline |
| [`drift/`](drift/) | Drift flag when rate moves >0.5pp vs baseline |
| [`inventory/`](inventory/) | Admin risk / webhook inventory checklist |
| [`staging/`](staging/) | Vendor re-verify + staging pass/fail |
| [`pack/`](pack/) | Agency / Plus pre-BFCM pack assembly |
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
