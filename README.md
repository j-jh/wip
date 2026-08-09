# wip

Learning project: thin Go wrappers around `dd-cli`, plus small demos you can run.

## Layout

```text
wip/
├── cmd/                      ← runnable demos (one idea each)
│   ├── list-addresses/       ← simplest read
│   └── reorder-preview/      ← phase-1 flow through quote (no charge)
├── internal/ddcli/           ← shared CLI adapter (start with client.go)
│   ├── client.go             ← RunCLICommand, BuildIntent, JSON envelope
│   ├── address.go
│   ├── cart.go
│   ├── order.go
│   └── stores.go
└── docs/
    ├── wip.md                ← what exists today
    ├── function-docs.md      ← comment / naming style (Go-learner friendly)
    └── plan/
        ├── tier-1-reads.md   ← early read-only roadmap (simple-read)
        └── phase-1.md        ← reorder → preview plan
```

Local-only (gitignored): `self-docs/`, `agent-ref/`, `.env`.

## Suggested reading order

1. `internal/ddcli/client.go` — how every command is run and decoded  
2. `internal/ddcli/address.go` — smallest wrapper  
3. `cmd/list-addresses` — smallest caller  
4. `internal/ddcli/order.go` + `cart.go` — reorder / preview  
5. `cmd/reorder-preview` — multi-step flow  

## Run demos

```bash
dd-cli login   # once
go run ./cmd/list-addresses
go run ./cmd/reorder-preview
```

Optional: set `DD_CLI_BIN` if `dd-cli` is not on your PATH.
