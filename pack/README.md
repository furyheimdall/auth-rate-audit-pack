# pack

Assembles `baseline/` + `drift/` + `inventory/` + `staging/` into **one scored report** (markdown + PDF stub) for agency SOW attach.

Other seats may still be stubs. This seat consumes **interfaces + fixtures** (`testdata/healthy.json`, `testdata/drift.json`) — no network.

```bash
go test ./pack
go run ./cmd/arap pack
```

See [docs/run.md](../docs/run.md) for how to run the pack and the locked IN/OUT reminders.

This directory is the report seat, not a SaaS landing page.
