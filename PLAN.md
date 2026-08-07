# Plan: Tier 1 dd-cli Go wrappers

Wrap the easiest **read-only** `dd-cli` commands as Go functions first. Scope is fewest params, no mutations, **no payment** until handling is decided later.

Source of truth: sibling notes `dd-cli-agent-reference.md` (v0.2.1). Always invoke with `--json-output`. Assume the machine already has a working `dd-cli login`.

---

## In scope (Tier 1)

| Go API (proposed) | CLI | Required flags | Optional flags |
|-------------------|-----|----------------|----------------|
| `AddressList` | `address list` | `--intent` | — |
| `CartList` | `cart list` | `--intent` | `--store-id` |
| `OrderHistory` | `order history` | `--intent` | `--max`, `--days` |
| `FindNearbyStores` | `find-nearby-stores` | `--intent` | `--vertical`, `--max`, `--lat`, `--lng` |

Excluded for now: `payment-method list`, all writes, `login`, Tier 2 ID lookups (`menu`, `cart show`, etc.).

---

## Shared plumbing (first)

1. **`internal/ddcli` package** — thin `exec` wrapper around the `dd-cli` binary.
2. **`Run(ctx, args ...string)`** — prepends `--json-output`, captures stdout/stderr, returns bytes + exit error.
3. **`Intent(summary, userPrompt string) string`** — builds the required intent blob once; every Tier 1 call takes an intent (or a prebuilt string).
4. **Binary resolution** — `DD_CLI_BIN` env override, else `dd-cli` on `PATH`.
5. **Response shape** — unmarshal into typed structs per command; keep a shared `Success` / `Message` envelope if present. Prefer decoding only fields we use next (IDs, names, labels).

Do not shell through a login helper in this phase.

---

## Package layout

```
wip/
├── PLAN.md                 ← this file
├── cmd/server/             ← later: HTTP surface over wrappers
└── internal/ddcli/
    ├── client.go           ← Run, Intent, Client config
    ├── address.go          ← AddressList
    ├── cart.go             ← CartList
    ├── order.go            ← OrderHistory
    └── stores.go           ← FindNearbyStores
```

Keep `cmd/server` empty/minimal until wrappers exist and can be called from tests or a small scratch main.

---

## Implementation order

1. **Client + Intent + Run** — one integration smoke: `dd-cli --version` or a dry parse of `address list` JSON when credentials exist.
2. **`AddressList`** — intent-only; validates end-to-end pattern (exec → JSON → struct).
3. **`CartList`** — same + optional `storeID *string` / empty = omit flag.
4. **`OrderHistory`** — optional `max` / `days` with CLI defaults when unset (`50` / `90`).
5. **`FindNearbyStores`** — optional vertical/max/lat/lng; default vertical left to CLI (`grocery`).
6. **Wire a tiny caller** — either table-driven tests (skip if no auth) or a `cmd/ddcli-probe` that prints one command’s JSON. Defer real HTTP in `cmd/server` until these four are stable.

---

## Function shape (convention)

```go
func (c *Client) AddressList(ctx context.Context, intent string) (*AddressListResult, error)

func (c *Client) CartList(ctx context.Context, intent string, storeID string) (*CartListResult, error)
// storeID == "" → omit --store-id

func (c *Client) OrderHistory(ctx context.Context, intent string, opts OrderHistoryOpts) (*OrderHistoryResult, error)

func (c *Client) FindNearbyStores(ctx context.Context, intent string, opts FindNearbyStoresOpts) (*FindNearbyStoresResult, error)
```

- Always pass `context.Context` for cancel/timeout.
- Omit optional CLI flags when zero/empty rather than sending defaults unless we need to override CLI defaults.
- Errors: wrap `exec` failures; if CLI returns JSON with `success: false`, surface `message` as a typed error when possible.

---

## Explicitly out of scope (this plan)

- `payment-method list` and any payment/submit/checkout path
- Mutations (`address set`, cart add/remove/delete, promo apply/remove, order submit/reorder)
- Commands needing `--items-json` or confirmation `--yes`
- Parsing every JSON field in the CLI envelope — only what Tier 1 consumers need
- Replacing `dd-cli` with direct DoorDash HTTP APIs

---

## Done when

- [x] Shared `Client.Run` + `Intent` + `AddressList` (first slice)
- [ ] `internal/ddcli` can run the four commands with `--json-output` + `--intent`
- [ ] Each returns a typed Go result (or clear error) without shelling out ad hoc elsewhere
- [ ] Optional flags are omit-when-empty
- [ ] No payment APIs in the package yet
- [x] Short note in `docs/wip.md` pointing at this plan and the package

---

## Next plan (not this one)

Tier 2 read-only by ID (`store-details`, `menu`, `cart show`, `order status`, `search`, …), then mutations, then payment handling strategy.
