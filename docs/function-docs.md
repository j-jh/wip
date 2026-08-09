# Function doc format

Use this shape for exported functions and important types. Plain language. Skip fluff.

**Audience:** someone learning Go. Prefer phrasing and code that stay simple to read: clear names, short steps, comments that explain *why* or unfamiliar Go ideas — not a line-by-line narration of obvious code.

```
## Name

One or two sentences: what it does / when to call it.

Params:
- name (type) — what it is; any empty/zero rules
- …

Returns:
- name (type) — what you get on success
- error — when it fails (and what that usually means)

Notes: (optional) CLI command, auth, side effects, omit if none
```

### Rules

- Always list every param and every return value.
- Say what empty strings / nil / zero mean when it matters.
- For wrappers, name the underlying `dd-cli` command in Notes.
- Types get a short field list instead of Params/Returns when they are data only.
- Prefer descriptive names in docs that match the Go identifiers (`ListDeliveryAddresses`, not vague shorthand).

### Writing for learners (code + comments)

When implementing or documenting wrappers, keep the **syntax, comments, and control flow** easy to follow:

- **Straight-line flow** — validate → run CLI → decode → check `success` → return. Avoid clever nesting, early-return puzzles, or dense one-liners.
- **Names over abbreviations** — spell out what a value is (`intentText`, `addressListResult`), not `it` / `res` / `o`.
- **Comments teach** — briefly explain Go or domain concepts a newcomer might not know (`context.Context` for cancel/timeout, why intent is required, what `structuredContent` is). Do not restate the next line of code.
- **One idea per step** — each block does one job so a reader can skim top to bottom without jumping.
- **Docs match the code** — param/return names in this comment shape should be the same identifiers used in the function signature.

### Example

```
## ListDeliveryAddresses

Lists the signed-in account’s saved delivery addresses.

Params:
- ctx (context.Context) — cancel / timeout for the CLI call
- intentText (string) — required DoorDash intent blob; use BuildIntent() to build it

Returns:
- *ListDeliveryAddressesResult — addresses plus success flag
- error — missing intent, CLI failure, bad JSON, or success:false from CLI

Notes: runs `dd-cli --json-output address list --intent …`. Needs prior `dd-cli login`.
```
