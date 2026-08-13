# wip

Plans:
- [plan/tier-1-reads.md](plan/tier-1-reads.md) — early read-only wrappers
- [plan/phase-1.md](plan/phase-1.md) — reorder → preview vertical slice

Function comment style: [function-docs.md](function-docs.md).

## Layout

- `internal/ddcli` — shared client + wrappers (`client.go` first)
- `cmd/list-addresses`, `cmd/reorder-preview` — tier-1 demos
- Manual tier-2 restaurant steps (not combined yet):
  - `cmd/search-restaurants`
  - `cmd/get-menu`
  - `cmd/list-open-carts`
  - `cmd/add-cart-items`
  - `cmd/preview-order`

## Implemented wrappers

`ListDeliveryAddresses`, `ListOrderHistory`, `FindNearbyStores`, `ListOpenCarts`, `ReorderPastOrder`, `PreviewOrder`, `SearchRestaurants`, `GetMenu`, `AddCartItems`
