# Plan: Tier 3 gated checkout

Place an order only after explicit human confirmation. Builds on payment-method list + restaurant → preview.

Branch: `feat/tier-3-gated-checkout`.

---

## In scope

| Go API | CLI | Notes |
|--------|-----|-------|
| `SubmitOrder` | `order submit --yes` | Wrapper always passes `--yes`; caller owns the gate |
| `GetOrderStatus` | `order status` | Poll until not `pending` |

| Demo | Purpose |
|------|---------|
| `cmd/submit-order` | Preview + payment summary → requires `-confirm-submit` (+ type `yes`) → charge |
| `cmd/order-status` | One-shot or `-poll` while pending |

---

## Gate rules

1. `WIP_ALLOW_SUBMIT=true` in `.env` (or process env). Default / example is `false`.
2. Demo always prints preview + payment label first.
3. Without `-confirm-submit`, exit with no charge.
4. With `-confirm-submit`, require typing `yes` unless `-yes` (scripted).
5. `-tip-cents` is required (cents, not dollars). Use `0` for pickup / no tip.

`SubmitOrder` itself refuses when the env flag is off — flags alone cannot charge.

---

## Out of scope

- Work-benefits / corporate budget flags
- `order checkout-url`
- Changing default payment method from CLI
- LLM / modifiers / grocery

---

## Checklist

- [x] `SubmitOrder` + `GetOrderStatus` wrappers
- [x] `cmd/submit-order` with confirm gate
- [x] `cmd/order-status` (+ optional poll)
- [x] README + docs updated
- [ ] Live charge test (optional; costs money)
