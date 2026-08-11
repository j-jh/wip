# Plan: Tier 1 read-only dd-cli wrappers

Wrap the easiest **read-only** `dd-cli` commands as Go functions first. Scope is fewest params, no mutations, **no payment** until handling is decided later.

Source of truth: sibling notes `dd-cli-agent-reference.md` (v0.2.1). Always invoke with `--json-output`. Assume the machine already has a working `dd-cli login`.

Branch: `feat/tier-1-simple-read`.

---

## In scope (Tier 1)

| Go API | CLI | Required flags | Optional flags |
|--------|-----|----------------|----------------|
| `ListDeliveryAddresses` | `address list` | `--intent` | — |
| `ListOpenCarts` | `cart list` | `--intent` | `--store-id` |
| `ListOrderHistory` | `order history` | `--intent` | `--max`, `--days` |
| `FindNearbyStores` | `find-nearby-stores` | `--intent` | `--vertical`, `--max`, `--lat`, `--lng` |

Excluded for now: `payment-method list`, all writes, `login`, ID lookups (`menu`, `cart show`, etc.).

---

## Shared plumbing (first)

1. **`internal/ddcli` package** — thin `exec` wrapper around the `dd-cli` binary.
2. **`RunCLICommand(ctx, args ...string)`** — prepends `--json-output`, captures stdout/stderr, returns bytes + exit error.
3. **`BuildIntent(summary, userPrompt string) string`** — builds the required intent blob once.
4. **Binary resolution** — `DD_CLI_BIN` env override, else `dd-cli` on `PATH`.
5. **Response shape** — decode outer envelope → `structuredContent` → typed result; treat `success: false` as an error.

Do not shell through a login helper in this phase.

---

## Package layout

```
wip/
├── docs/plan/tier-1-reads.md     ← this file
├── cmd/list-addresses/           ← simplest demo caller
└── internal/ddcli/
    ├── client.go                 ← start here
    └── address.go                ← ListDeliveryAddresses (done)
    # later on this track / other branches:
    # cart.go, order.go, stores.go
```

---

## Implementation order

1. **Client + BuildIntent + RunCLICommand** — shared pattern.
2. **`ListDeliveryAddresses`** — intent-only; proves exec → JSON → struct.
3. **`ListOpenCarts` / `ListOrderHistory` / `FindNearbyStores`** — same pattern + optional flags (other branch / next slices).
4. Keep demos as `cmd/<name>` with one idea each (not a fake “server”).

---

## Function shape (convention)

```go
func (client *CLIClient) ListDeliveryAddresses(ctx context.Context, intentText string) (*ListDeliveryAddressesResult, error)
```

- Always pass `context.Context` for cancel/timeout.
- Omit optional CLI flags when zero/empty.
- Errors: wrap `exec` failures; if CLI JSON has `success: false`, surface `message`.

Docs comments follow [docs/function-docs.md](../function-docs.md) (Go-learner friendly).

---

## Explicitly out of scope (this plan)

- `payment-method list` and any payment/submit/checkout path
- Mutations (`address set`, cart add/remove/delete, promo, order submit/reorder)
- Commands needing `--items-json` or confirmation `--yes`
- Parsing every JSON field — only what callers need next
- Replacing `dd-cli` with direct DoorDash HTTP APIs

---

## Done when

- [x] Shared `CLIClient` + `BuildIntent` + `RunCLICommand` + `ListDeliveryAddresses`
- [x] `cmd/list-addresses` demo
- [ ] Remaining Tier 1 reads (`ListOpenCarts`, `ListOrderHistory`, `FindNearbyStores`) — see `feat/tier-1-reorder` for several of these
- [x] Optional-flag convention documented
- [x] No payment APIs in the package yet
- [x] Docs under `docs/` with a clear reading order in `README.md`

---

## Next plan (not this one)

Reorder → preview vertical slice, then checkout, then restaurant/grocery cart builders.
