# wip

See [plan/tier-1-reads.md](plan/tier-1-reads.md) for the early read-only wrapper roadmap.

Function comment style: [function-docs.md](function-docs.md).

## Layout

- `internal/ddcli` — shared client + wrappers (`client.go` first, then `address.go`)
- `cmd/list-addresses` — simplest demo (address list)

## Implemented wrappers

`ListDeliveryAddresses` (plus shared `CLIClient`, `BuildIntent`, `RunCLICommand`)
