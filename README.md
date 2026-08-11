# wip

Learning project: thin Go wrappers around `dd-cli`, plus small demos you can run.

## Layout

```text
wip/
├── cmd/
│   └── list-addresses/       ← simplest read demo
├── internal/ddcli/           ← shared CLI adapter (start with client.go)
│   ├── client.go             ← RunCLICommand, BuildIntent, JSON envelope
│   └── address.go            ← ListDeliveryAddresses
└── docs/
    ├── wip.md                ← what exists today
    ├── function-docs.md      ← comment / naming style (Go-learner friendly)
    └── plan/
        └── tier-1-reads.md   ← early read-only roadmap
```

Local-only (gitignored): `self-docs/`, `agent-ref/`, `.env`.

## Suggested reading order

1. `internal/ddcli/client.go` — how every command is run and decoded  
2. `internal/ddcli/address.go` — smallest wrapper  
3. `cmd/list-addresses` — smallest caller  

## Run demo

```bash
dd-cli login   # once
go run ./cmd/list-addresses
```

Optional: set `DD_CLI_BIN` if `dd-cli` is not on your PATH.
