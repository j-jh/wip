# Plan: Tier 3 checkout read-only

First checkout slice after Tier 2 restaurant → preview. **Read-only only** — no submit, no charge.

Branch: `feat/tier-3-checkout-readonly`.

---

## In scope (this branch)

| Go API | CLI | Notes |
|--------|-----|-------|
| `ListPaymentMethods` | `payment-method list` | Cards on file + `default_payment_method_id` |

| Demo | Purpose |
|------|---------|
| `cmd/list-payment-methods` | Run after preview to surface saved cards (masked) |

---

## Out of scope (later branch)

- `order submit`, `order status`, `order checkout-url`
- Gated `--confirm-submit` flow
- `GetRestaurantItemDetails` / modifiers

---

## Checklist

- [x] `ListPaymentMethods` wrapper in `internal/ddcli/payment.go`
- [x] `cmd/list-payment-methods` demo
- [x] README + `docs/wip.md` updated
- [ ] Manual test after `restaurant-preview` (logged-in account)

---

## Notes

`payment-method list` returns **cards only** — wallets (Apple Pay, etc.) are not in `cards[]`. Empty list does not mean no payment methods on file. See `dd-cli payment-method list --help`.
