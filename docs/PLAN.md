---
PLAN: "refactor(components): drop author-invented element ids for Key + Ref()"
TAG: v0.6.19
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 18267888995124046778
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **`webtyp.com/dom v0.13.11` is published and already required by this repo's
> `go.mod`.** It is what introduces `(*Element).Ref()` and makes a keyed element
> addressable; every stage below depends on it. Do not downgrade it.

# PLAN — components stop naming element ids

## Why

`dom` now offers the typed contract that was missing: an author keeps the
`*Element` they built, marks it with `Key()`, and asks it for its live node with
`Ref()`. No global name is invented, so two instances of one component on the
same page cannot collide.

```go
// before — a name the author made up, global, collides at the second instance
row := Li().ID("row-"+key)
ref, ok := Get("row-" + key)

// after — a handle, scoped to the instance that built it
row := Li().Key(key)
ref, ok := row.Ref()
```

This was already proved in `calendarslider` (v0.6.18), which had to drop
`id="cs-m-…"` because two calendars in `app-demo` panicked `dom.claimID`. The
rest of the package still carries the old pattern.

**`ID()` is not forbidden** and this plan does not add any check. It stays valid
for elements the *application or chassis* declares outside a component. What
changes is that components stop using it.

## Rules for this plan

- **Never invent a replacement upstream symbol.** If a site needs something
  `dom` does not expose (e.g. an `aria-describedby` equivalent of
  `(*Element).For`), **stop and report it** in the PR description — do not
  declare a local helper. Per `CONSTRUCTION_HARNESS.md`: *"A missing contract at
  a boundary is a defect in the library, not in the consumer."*
- **No `map`** (TinyGo binary budget) and **no standard library** — use
  `webtyp/fmt`.
- Keep every `Key()` that exists: `BindChildren` reconciles on it.
- `gotest ./...` must stay green, `wasm ✅` included.

## Stage 1 — the three list components: delete a redundant line

`targetlist`, `targetdate` and `targethour` each build their row as
`ID(key).Key(key)` — the same string twice. With dom v0.13.11 the `Key` alone
makes the row addressable, so the `ID` line is pure duplication **and** the
thing that collides when two lists render the same key.

Delete the `ID(key).` line only. Change nothing else.

| File | Line |
|---|---|
| `targetlist/targetlist.go` | 145 |
| `targetdate/targetdate.go` | 146 |
| `targethour/targethour.go` | 161 |

Each becomes:

```go
	row := Li().Set(clsRow.AsAttr()).
		Key(key).
		…
```

Then `grep -rn 'ID(key)' targetlist/ targetdate/ targethour/` → **empty**.

## Stage 2 — `usermenu`: `Ref()` instead of `Get(GetID())`

`usermenu/usermenu.go:148` and `:154` use the old workaround
`Get(menu.GetID())`. That call silently returns `false` for any element `dom`
never id'd; `Ref()` is the contract for exactly this.

```go
	menu.On("toggle", func(Event) {
		if ref, ok := menu.Ref(); ok {
			m.open.Set(ref.GetAttr("open") != "<null>")
		}
	})

	backdrop.On("click", func(Event) {
		if ref, ok := menu.Ref(); ok {
			ref.RemoveAttr("open")
		}
		m.open.Set(false)
	})
```

`menu` already carries an event, so it is already id'd; no `Key` needed here.

## Stage 3 — `calendarslider`: anchor on the month card again

v0.6.18 had to anchor `slideToMonth` on the month's `‹` **button** because the
month card itself carried no id (it has no events or bindings). With
v0.13.11 the card's existing `Key(key)` makes it addressable, so the anchor goes
back to the card — what the code always meant.

In `calendarslider/calendarslider.go`:

- `monthRef` becomes `{ key string; el *Element }` and its doc comment drops the
  "‹ button" explanation, saying instead: the card is addressable because it
  carries `Key(key)`.
- `buildMonth` records the card: `c.months = append(c.months, monthRef{key: key, el: monthEl})`.
- `slideToMonth` resolves `el.Ref()` instead of `Get(anchor.GetID())`:

```go
	ref, ok := el.Ref()
	if !ok {
		return
	}
```

`data-month` stays — the tests select on it, and it is a breadcrumb, not an id.

## Stage 4 — `sitenav`, `scheduleeditor`, `selectsearch`

These three invent names that would collide if the component were mounted twice.

**`sitenav/sitenav.go:106`** — `menu := Div().Set(clsNavMenu.AsAttr()).ID(menuID)`.
Replace `ID(menuID)` with `Key(menuID)` and resolve through `Ref()` wherever
`menuID` is currently looked up. Grep `menuID` in that file first and migrate
every use together; if `menuID` is also referenced from a `<label for>` or an
`aria-*` attribute, see the note in Stage 5.

**`scheduleeditor/scheduleeditor.go:352`** — a radio group where
`id := "scheduleeditor-type-" + opt.Key` pairs each `<input type=radio>` with its
`<label for=…>`. Two editors on one page would collide on every option id.
The label/input pairing has a typed path already: **`(*Element).For(other)`**,
which points `for=` at the other element's dom-assigned id. Build the radio
first, then the label:

```go
		radio := Input("radio").Set(clsExcType.AsAttr()).
			Key(opt.Key).
			Attr("name", "scheduleeditor-type").
			…
		label := Label().For(radio).…
```

Keep `Attr("name", …)` as it is — the radio group name is an HTML grouping
concept, not an id, and grouping two editors under one name is a separate
question that is **out of scope here**.

**`selectsearch/selectsearch.go`** — four sites (lines 214, 265, 294, 369) built
from a per-instance `c.uid` prefix, the same scheme `calendarslider` deleted in
v0.6.18:

| Line | Element | Migration |
|---|---|---|
| 214 | `toggle` checkbox, `ID(c.uid+suffixToggle)` | paired with a `<label>`: nest the input inside the label (implicit association, no `for`/`id`) or use `For(toggle)` — whichever the existing markup allows |
| 265 | `searchInput`, `ID(c.uid+suffixSearch)` | `Key(suffixSearch)` + `Ref()` at the use sites |
| 294 | `optList`, `ID(c.uid+suffixOptions)` | `Key(suffixOptions)` + `Ref()` |
| 369 | `item`, `ID(c.uid+suffixOption+opt.ID)` | already has `Key(opt.ID)` — the comment says the id is *"required for wirePendingEvents to attach the click handler"*, which dom now does for any element carrying an event; **verify with a WASM test that the click still fires, then delete the `ID` line** |

Once all four are gone, delete `c.uid`, its generator, and the `suffix*`
constants that no longer have a reader — `grep -rn 'uid' selectsearch/` must come
back empty of the id scheme.

## Stage 5 — if an id is genuinely required by HTML semantics

`<label for>`, `aria-describedby`, `aria-labelledby` and `<datalist>` need a real
id. Use `(*Element).For(other)` where the relation is label→control. **If the
relation is an `aria-*` one and `dom` exposes no equivalent of `For`, do not
build one here** — record it in the PR description as a `dom` gap. That is the
finding, not a defect of this plan.

## Tests

- Existing tests keep passing; update only selectors that referenced a deleted
  id (prefer `data-*` attributes or classes, as `calendarslider` did).
- **One new WASM test per component touched in Stages 3-4**, shaped like
  `calendarslider`'s `TestTwoInstancesShareAPage`: mount TWO instances as
  siblings in one render, assert no panic and that acting on A does not affect B.
  That test is the proof the id collision is gone.

## Acceptance criteria

- `grep -rn '\.ID(\|^\s*ID(' --include=*.go . | grep -v _test | grep -v '/web/'`
  → **empty** (demo `main`s under `web/` legitimately keep `ID("app-result")`).
- `grep -rn 'Get(.*GetID())' --include=*.go .` → **empty**.
- `gotest ./...` green, `wasm ✅` and `race ✅` included.
- `gofmt -l .` → empty; `go vet ./...` clean.

## Out of scope

- `webtyp/dom` — it is the upstream and already published.
- `webtyp/form`, `webtyp/layout` — their own plans.
- Any check or panic forbidding `ID()`. That is a later `dom` plan, only after
  every consumer has migrated.

## Stages

| # | Files | Change |
|---|---|---|
| 1 | `targetlist`, `targetdate`, `targethour` | delete the redundant `ID(key)` line |
| 2 | `usermenu/usermenu.go` | `Get(menu.GetID())` → `menu.Ref()` (×2) |
| 3 | `calendarslider/calendarslider.go` | anchor back on the month card via `Ref()` |
| 4 | `sitenav`, `scheduleeditor`, `selectsearch` | `Key` + `Ref()` / `For()`; delete the `uid` scheme |
| 5 | — | report any missing `aria-*` contract upstream instead of patching |
