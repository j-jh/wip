# wip

Plans:
- [plan/tier-1-reads.md](plan/tier-1-reads.md) — early read-only wrappers (kept from simple-read)
- [plan/phase-1.md](plan/phase-1.md) — reorder → preview vertical slice

Function comment style: [function-docs.md](function-docs.md).

## Layout

- `internal/ddcli` — shared client + wrappers (`client.go` first, then command files)
- `cmd/list-addresses` — simplest demo (address list)
- `cmd/reorder-preview` — history → cart check → reorder → preview (stops; no submit)

## Implemented wrappers

`ListDeliveryAddresses`, `ListOrderHistory`, `FindNearbyStores`, `ListOpenCarts`, `ReorderPastOrder`, `PreviewOrder`
