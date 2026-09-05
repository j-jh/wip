# wip

Plans:
- [plan/tier-1-reads.md](plan/tier-1-reads.md) — early read-only wrappers
- [plan/phase-1.md](plan/phase-1.md) — reorder → preview vertical slice
- [plan/tier-3-checkout-readonly.md](plan/tier-3-checkout-readonly.md) — payment-method list
- [plan/tier-3-gated-checkout.md](plan/tier-3-gated-checkout.md) — gated submit + status

Function comment style: [function-docs.md](function-docs.md).

## Layout

- `internal/ddcli` — shared client + wrappers (`client.go` first)
- Tier 1: `cmd/list-addresses`, `cmd/reorder-preview`
- Tier 2 restaurant (manual + combined):
  - `cmd/search-restaurants`, `cmd/get-menu`, `cmd/list-open-carts`
  - `cmd/add-cart-items`, `cmd/preview-order`, `cmd/delete-cart`
  - `cmd/restaurant-preview` — interactive login → search → menu → add → preview
- Tier 3: `cmd/list-payment-methods`, `cmd/submit-order`, `cmd/order-status`

## Implemented wrappers

`ListDeliveryAddresses`, `ListOrderHistory`, `FindNearbyStores`, `ListOpenCarts`, `ReorderPastOrder`, `PreviewOrder`, `SearchRestaurants`, `GetMenu`, `AddCartItems`, `DeleteCart`, `ListPaymentMethods`, `SubmitOrder`, `GetOrderStatus`
