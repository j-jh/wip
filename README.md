# wip

Rough Plan:

Problem statements: 
1. What's the balance between agent automation and human intervention?
2. How do we provide a non chatbox based UI?
3. When does the agent ask vs assume?
4. What separates this from another llm wrapper over a traditional backend?

Deterministic:
1. Backend wrappers over dd-cli commands
    a. Functions for:
        a. Search for... select restaurant 
        b. Add to cart (req. modifiers, notes)
        c. Checkout/payment
        d. 
2. Human prompting interactions for each above dd based flow step (no llm)

Agentified:
3. Gradually introduce llm... (AI takeover)
    a. Search flow: find me (cuisine/category) 
        1. Agent decides where to order?
    b. Single item flow: add item with required and optional modifiers + notes to my cart
        1. Agent decides what to add/modify?
    c. Ordering flow: repeat b. for given list of items, checkout 
        1. Agent decides the whole cart state?

What must/must not be automated? Full intent understanding, tool routing...Agentify process? Memory? User habits and recommendations?

Human determinism reduces as we progress down this chain.

Reorder flow: low intervention, stable demoable path to checkout.

New restaraunt: high human in the loop intervention by design until LLM implementation

-----

Learning project: thin Go wrappers around `dd-cli`, plus small demos you can run.

## Progress

| Stage | Status | Demos / wrappers |
|-------|--------|------------------|
| **Tier 1** — reorder path | Done | `list-addresses`, `reorder-preview` |
| **Tier 2** — restaurant → preview | Done | `search-restaurants` … `preview-order`, `delete-cart`, `restaurant-preview` |
| **Tier 3** — payment list | Done | `list-payment-methods` (`ListPaymentMethods`) |
| **Tier 3** — gated submit | Done | `submit-order` (`-confirm-submit`), `order-status` |
| **Later** | — | item modifiers, grocery, LLM tool routing |

Detailed plans: `docs/plan/`.

## Layout

```text
wip/
├── cmd/                      ← runnable demos (one idea each)
│   ├── list-addresses/
│   ├── reorder-preview/      ← combined reorder → preview
│   ├── search-restaurants/   ← tier-2 manual step
│   ├── get-menu/
│   ├── list-open-carts/
│   ├── add-cart-items/
│   ├── preview-order/
│   ├── delete-cart/
│   ├── restaurant-preview/   ← interactive combined flow (prompts)
│   ├── list-payment-methods/ ← tier-3: saved cards (read-only)
│   ├── submit-order/         ← tier-3: gated charge (needs -confirm-submit)
│   └── order-status/         ← tier-3: poll after submit
├── internal/ddcli/           ← shared CLI adapter (start with client.go)
└── docs/
    ├── wip.md
    ├── function-docs.md
    └── plan/
```

Local-only (gitignored): `self-docs/`, `agent-ref/`, `.env`.

## Prerequisites

```bash
dd-cli login          # once, and again when the token expires
go version            # Go 1.22+ (see go.mod)
```

Optional: `export DD_CLI_BIN=/path/to/dd-cli` if the binary is not on your `PATH`.

## How to run

All commands are run from the repo root.

### Tier 1 (already combined)

```bash
go run ./cmd/list-addresses
go run ./cmd/reorder-preview    # history → cart check → reorder → preview (no charge)
```

### Tier 2 restaurant → preview (interactive combined)

Prompts at each step; stops after quote (no charge). Open-cart reuse is inside `AddCartItems`.

```bash
go run ./cmd/restaurant-preview
# typical prompts: query, max results, lat,lng (e.g. 37.76,-122.48), store #, item #, qty, pickup
```

### Tier 2 restaurant → preview (manual steps)

Run **one command at a time**. Copy IDs from each command’s output into the next flags. Do not submit/pay here.

**1. Search restaurants** → note a `store_id`

```bash
go run ./cmd/search-restaurants -query "thai" -limit 5

# optional location override (both required together):
go run ./cmd/search-restaurants -query "thai" -limit 5 -lat 37.76 -lng -122.48
```

**2. Get menu** → note `menu_id` and an `item_id` / name (prefer a simple item)

```bash
go run ./cmd/get-menu -store-id STORE_ID
```

**3. List open carts** → check for an existing cart at that store (one open cart per store)

```bash
go run ./cmd/list-open-carts
go run ./cmd/list-open-carts -store-id STORE_ID
```

**4. Add one item** → note `cart_uuid`

```bash
go run ./cmd/add-cart-items \
  -store-id STORE_ID \
  -menu-id MENU_ID \
  -item-id ITEM_ID \
  -item-name "Item Name" \
  -quantity 1

# optional: append to an existing cart, or set fulfillment on create
go run ./cmd/add-cart-items \
  -store-id STORE_ID -menu-id MENU_ID \
  -item-id ITEM_ID -item-name "Item Name" -quantity 1 \
  -cart-uuid CART_UUID \
  -fulfillment pickup
```

**5. Preview quote** (no charge)

```bash
go run ./cmd/preview-order -cart-uuid CART_UUID
```

**6. Abandon cart** (optional)

```bash
go run ./cmd/delete-cart -cart-uuid CART_UUID
```

### Tier 3 checkout read-only

After preview (or anytime when logged in). Cards only — no charge.

```bash
go run ./cmd/list-payment-methods
```

### Tier 3 gated submit (charges money)

Shows preview + payment summary first. Without `-confirm-submit`, exits with **no charge**.
With `-confirm-submit`, type `yes` (or pass `-yes` to skip typing). Tip is in **cents**.

```bash
# dry run — prints quote, does not charge
go run ./cmd/submit-order -cart-uuid CART_UUID -tip-cents 0

# place order (destructive)
go run ./cmd/submit-order -cart-uuid CART_UUID -tip-cents 0 -confirm-submit

# after submit
go run ./cmd/order-status -order-uuid ORDER_UUID
go run ./cmd/order-status -order-uuid ORDER_UUID -poll
```

### Flag cheat sheet

| Command | Required flags | Optional |
|---------|----------------|----------|
| `search-restaurants` | `-query` | `-limit`, `-lat`+`-lng` |
| `get-menu` | `-store-id` | — |
| `list-open-carts` | — | `-store-id` |
| `add-cart-items` | `-store-id`, `-menu-id`, `-item-id`, `-item-name` | `-quantity`, `-cart-uuid`, `-fulfillment` |
| `preview-order` | `-cart-uuid` | — |
| `delete-cart` | `-cart-uuid` | — |
| `list-payment-methods` | — | — |
| `submit-order` | `-cart-uuid`, `-tip-cents` | `-confirm-submit`, `-yes`, `-fulfillment` |
| `order-status` | `-order-uuid` | `-poll` |

## Suggested reading order

1. `internal/ddcli/client.go`  
2. `internal/ddcli/address.go` + `cmd/list-addresses`  
3. Tier 2: `search.go` → `menu.go` → `cart.go` → `order.go` (`PreviewOrder`)  
4. Tier 3 read-only: `payment.go` + `cmd/list-payment-methods`  
5. Tier 3 gated: `order.go` (`SubmitOrder`, `GetOrderStatus`) + `cmd/submit-order`
