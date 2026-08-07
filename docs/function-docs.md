# Function doc format

Use this shape for exported functions and important types. Plain language. Skip fluff.

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

### Example

```
## AddressList

Lists the signed-in account’s saved delivery addresses.

Params:
- ctx (context.Context) — cancel / timeout for the CLI call
- intent (string) — required DoorDash intent blob; use Intent() to build it

Returns:
- *AddressListResult — addresses plus success flag
- error — missing intent, CLI failure, bad JSON, or success:false from CLI

Notes: runs `dd-cli --json-output address list --intent …`. Needs prior `dd-cli login`.
```
