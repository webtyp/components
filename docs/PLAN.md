---
PLAN: "refactor(css): every button goes through style.Button — one recipe, nine sites"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 11090955321065316234
PR: https://github.com/webtyp/components/pull/28
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Its gate is already shipped: **`webtyp.com/widget v0.6.26`** exposes
> `style.Button(Surface)`. Orchestrator:
> [webtyp/docs/BUTTON_SYSTEM_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/BUTTON_SYSTEM_MASTER_PLAN.md).

# Plan — every button in this repo goes through `style.Button`

## 1. Context you need (zero assumptions about this repo)

`webtyp.com/components` is a library of pure UI components. Each component
declares its CSS in a `css.go` file guarded by `//go:build !wasm`, built with
the `webtyp.com/widget/style` DSL:

```go
func (b *ActionButton) RenderCSS() *css.Stylesheet {
	return style.For(b).
		Root(style.Pad(style.Space2)).
		Part(PartPrimary, style.Interactive(style.Primary)).
		Stylesheet()
}
```

`style` offers **box recipes** (`ControlBox()`, `IconBox()`, `ChipBox()`,
`LogoBox()`) and a **paint recipe**, `Interactive(Surface)`. Until
`widget v0.6.26` there was no recipe for a button, so nine components composed
their own — and they disagreed. Two shipped defects, measured live in the demo:
`.scheduleeditor__row-add` rendered **799.16 px wide** (a gradient bar across
its panel), and `.scheduleeditor__row-remove` overflowed its own box.

`widget v0.6.26` adds the missing member:

```go
// Button is the one recipe for a button: the shared control height, inline
// padding for its label, and a box that neither stretches to its container
// nor shrinks under pressure. s paints it with the hover, focus and press
// treatments Interactive derives, and with that surface's default radius.
func Button(s Surface) Option
```

It emits `min-height: var(--control-height)`, `padding-inline: var(--space-3)`,
`align-self: center`, `flex-shrink: 0`, `flex-grow: 0`, plus the interactive
surface. It deliberately emits **no `display`** and **no `border-radius`** — a
native `<button>` centres its own label, and the surface layer already applies
`defaultRadius()`.

`Sheet.Validate()` in `v0.6.26` now **rejects** two compositions:

- `Button()` beside `Interactive()` → `Button already paints an interactive surface; drop Interactive`
- `Button()` beside `ControlBox()` → `Button already carries the control height; drop ControlBox`

So a half-done migration fails validation rather than shipping silently.

## 2. The rule this plan applies

> **`Button()` is for what the user presses.** `Interactive()` stays for an
> interactive surface that is not a button — it paints without claiming a
> control's box.

Every site below is already classified. **Do not reclassify anything**, and do
not migrate a site this plan does not list. Both lists are exhaustive.

## 3. Stage 1 — bump the dependency

From the repo root:

```
go get webtyp.com/widget@v0.6.26
go mod tidy
```

`go.mod` must end up requiring `webtyp.com/widget v0.6.26` or higher. Do not
bump any other dependency in this plan.

**Careful with `replace webtyp.com/icons => ../icons`.** If `go mod tidy`
removes that line, the `require` behind it falls back to the placeholder
`v0.0.3` it was pinned at — four tags behind the published `v0.0.7`, a silent
downgrade. Either leave the replace alone, or drop it **and** raise the require
to `webtyp.com/icons@v0.0.7`. Never leave `v0.0.3` requireless of a replace.

## 4. Stage 2 — migrate the nine button sites

For each site: replace the hand-rolled recipe with a single `style.Button(s)`,
keeping the same `Surface` argument the site already used.

**Delete, at each migrated site,** every Option that `Button` now owns:
`style.Interactive(...)`, `style.ControlBox()`, `style.KeepSize()`, and
`style.Round(style.RadiusSm)`. Leave every other Option on the part untouched
(`Row`, `Pad`, `FontSize`, `Anchor`, `Animate`, `Grow`, …) unless a stage below
says otherwise.

**`Round(style.RadiusSm)` and only that one.** `Button` derives its surface's
default radius, and `Primary`/`Secondary`/`Danger`/`Subtle` all resolve to
`RadiusSm` — so that call is the redundant one. A part carrying a *different*
radius chose a shape deliberately and **keeps it**: `themetoggle`'s root and
`usermenu`'s `PartTrigger` both carry `Round(style.RadiusFull)`, which is what
makes them a circle and a pill. Deleting those silently turns both into
4px-cornered boxes.

### 4.1 `actionbutton/css.go`

Today the root carries the box and the three parts only paint:

```go
		Root(
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
			style.As(style.Page),
		).
		Part(PartPrimary,
			style.Interactive(style.Primary),
		).
		Part(PartSecondary,
			style.Interactive(style.Secondary),
		).
		Part(PartDanger,
			style.Interactive(style.Danger),
		).
```

Becomes — the box moves into each part, because that is what `Button` is, and
the root keeps nothing that restates it:

```go
		Part(PartPrimary,
			style.Button(style.Primary),
		).
		Part(PartSecondary,
			style.Button(style.Secondary),
		).
		Part(PartDanger,
			style.Button(style.Danger),
		).
```

`Pad(Space2)` and `Round(RadiusSm)` are now `Button`'s job, and `As(style.Page)`
under three interactive parts painted a surface no pixel of the widget ever
showed — all three go.

**The `Root(...)` call itself must stay, with a real rule.** The markup is
`class="actionbutton actionbutton__primary"` — root class and variant class on
the same node — so `TestPairMarkupAndStylesheet` fails with *"HTML class
\"actionbutton\" is unstyled in CSS"* if the root ends up ruleless. Do not
satisfy it with filler. The rule that belongs there is the one the variants
cannot supply:

```go
		// The root is the pressed element itself — the markup is
		// class="actionbutton actionbutton__primary", both rules on one node —
		// so it may not be ruleless (TestPairMarkupAndStylesheet). What belongs
		// here is the one thing the variants cannot supply: with Href set this
		// renders an <a>, which unlike a native <button> does not centre its own
		// label inside the padding Button() adds.
		Root(
			style.CenterContent(),
		).
```

### 4.2 `scheduleeditor/css.go` — four sites

| Part | Today | Becomes |
|---|---|---|
| `PartRowRemove` | `ControlBox(), Interactive(Danger), Round(RadiusSm), KeepSize()` | `Button(style.Danger)` |
| `PartRowAdd` | `ControlBox(), Interactive(Primary), Round(RadiusSm), KeepSize()` | `Button(style.Primary)` |
| `PartExcAdd` | `ControlBox(), Interactive(Primary), Round(RadiusSm), KeepSize()` | `Button(style.Primary)` |
| `PartExcRemove` | `ControlBox(), Interactive(Danger), Round(RadiusSm), KeepSize()` | `Button(style.Danger)` |

`PartRowAdd` is the 799px bar; after this it takes its content width.

**Do NOT touch `PartDayChip`** in this file. It is a native `<input
type="checkbox">` and its `ControlBox()` is deliberate — `css.ControlWidth`'s
own comment says it exists so a checkbox's tap target meets the 44px touch
floor on both axes. It is not a button and it is not in scope. See §8.

### 4.3 `themetoggle/css.go`

The root is the `<button>` itself. Replace `style.Interactive(style.Primary)`
with `style.Button(style.Primary)`, and delete any `ControlBox()`,
`KeepSize()` or `Round(style.RadiusSm)` on that same `Root(...)` call. Keep
everything else.

### 4.4 `usermenu/css.go`

`PartTrigger` is a `<summary>` acting as a button. Replace
`style.Interactive(style.Subtle)` with `style.Button(style.Subtle)`, deleting
`ControlBox()`, `KeepSize()` and `Round(style.RadiusSm)` from that part if
present. Keep its layout Options (`Row`, `CenterContent`, …).

## 5. Stage 3 — the eight sites that must NOT change

Read this list, change nothing in it, and do not "finish the job" by migrating
them. Each is an interactive surface that is not a button:

| File | Part | Why it stays |
|---|---|---|
| `calendarslider/css.go` | `PartPrev`, `PartNext` | square icon controls sized by `IconBox(IconLg)`; the file's own comment explains they scale globally through the IconSize scale. `Button`'s `padding-inline` would break the square. |
| `calendarslider/css.go` | `PartCollapsed` | deliberately a full-width bottom bar (its comment says so); `Button` would shrink it to content width |
| `calendarslider/css.go` | `PartDaySelectable` | a calendar cell |
| `datatable/css.go` | `PartRow` | a table row |
| `selectsearch/css.go` | `PartOption` | a list option |
| `targethour/css.go` | `PartRow`, `PartFree` | list rows |
| `targetlist/css.go` | `PartRow` | a list row |
| `targetdate/css.go` | `PartRow` | a list row |

## 6. Stage 4 — the conformance guard

**File: [`conformance_test.go`](../conformance_test.go).**

Without a guard the next component re-invents its own button and the whole
migration erodes. This repo already has the mechanism: `TestNoRemovedSymbols`
walks every `css.go` and rejects forbidden substrings. Add a sibling test
following **exactly** that function's shape — same `filepath.Walk(".")`, same
`docs`/`.git`/`web` `SkipDir`, same `strings.HasSuffix(path, "css.go")` filter,
same `os.ReadFile`:

```go
// A button is style.Button(Surface) and nothing else. Composing one by hand
// from ControlBox + Interactive + Round + KeepSize is what left nine
// components disagreeing about a button's height and shipped an 800px-wide
// "Add row" bar. If a part needs an interactive surface that is NOT a button —
// a table row, a calendar cell, a list option — it keeps Interactive() and
// does not carry ControlBox() alongside it.
func TestNoHandRolledButtons(t *testing.T) {
```

The check, per `css.go`: for every part rule that contains **both**
`style.Interactive(` and `style.ControlBox(`, report

`%s: hand-rolled button (Interactive + ControlBox); use style.Button(Surface)`

Split the file into part blocks on the literal `\t\tPart(` and `\t\tRoot(`
boundaries and test each block, so a file that legitimately has an
`Interactive` part and a separate `ControlBox` part does not trip the check.

Four files in §5 legitimately pair `Interactive` with `ControlBox` on one part
(`selectsearch/PartOption`, `targethour/PartRow` and `PartFree`,
`targetlist/PartRow`, `targetdate/PartRow`). Add them to an explicit allowlist
inside the test, as a `map[string][]string` of file → part names, each entry
carrying a one-line comment saying why that part is a row and not a button. An
allowlist that must be edited deliberately is the point: it makes the next
addition a decision instead of an accident.

## 7. Stage 5 — verify

```
gotest ./...
go vet ./... && gofmt -l .
```

All green. `gotest` runs `Sheet.Validate()` through the existing
`TestConformance`/`TestEveryPackageEmits` cases, so a half-migrated part
(`Button` still beside `Interactive`) fails there with the `widget` message
quoted in §1 — that is the intended safety net, not a surprise.

## 8. Out of scope — do not attempt

The `scheduleeditor` **day chips** render as native 44×50 checkboxes. The size
is correct (touch floor). What reads badly is the native checkbox skin, and
making them read as pills needs `appearance: none`, which `widget/style` does
not expose — `grep -rn 'appearance' style/` in `widget` returns nothing.

**Do not work around it here**: no local CSS string, no hiding the input behind
a styled label, no new `Part` for the chip. A missing primitive is fixed
upstream in `widget` under its own plan, never re-created in a consumer. Leave
`PartDayChip` exactly as it is.

## 9. Constraints

- `gotest`, never `go test`. Never run `gopush` or `codejob` — both are the
  developer's tools, outside this plan.
- No standard library in WASM-compiled packages: use `webtyp.com/fmt` instead of
  `errors`, `strconv`, `strings`. **`conformance_test.go` is a `!wasm` test file
  and already imports `os`, `strings`, `path/filepath`, `regexp` — that is
  correct and legitimate; do NOT "fix" those imports.**
- CSS lives only in `css.go` files with `//go:build !wasm`. Never move a rule
  out of one.
- No `TODO`, no `FIXME`, no deprecated path, no commented-out old recipe left
  beside the new one. Before closing:
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` and confirm every hit
  predates this change.
- Do not change any component's rendered HTML, its `Part` constants, or its
  public Go API. This plan touches `css.go` files, `go.mod`, `go.sum` and
  `conformance_test.go` — nothing else.

## 10. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `grep -rn 'style.Button(' --include=css.go . \| wc -l` | **9** |
| 2 | `grep -rn 'style.Interactive(' --include=css.go . \| wc -l` | **11** — the §5 list (4 in `calendarslider`) plus `scheduleeditor`'s `PartDayChip` |
| 3 | `grep -n 'Interactive\|ControlBox\|KeepSize\|Round' scheduleeditor/css.go` | no hit on `PartRowAdd`, `PartRowRemove`, `PartExcAdd`, `PartExcRemove` |
| 4 | `grep -n 'PartDayChip' -A 4 scheduleeditor/css.go` | unchanged: still `ControlBox()`, `Round(RadiusSm)`, `Interactive(Subtle)` |
| 5 | `grep -n -A 2 'Root(' actionbutton/css.go` | `style.CenterContent()` — one real rule, no `Pad`/`Round`/`As` |
| 6 | `grep -n 'webtyp.com/widget v0.6.2' go.mod` | `v0.6.26` or higher |
| 6b | `grep -n 'icons' go.mod` | `v0.0.7`, or a `replace` — never a bare `v0.0.3` |
| 6c | `grep -n -B 3 'RadiusFull' themetoggle/css.go usermenu/css.go` | present on `themetoggle`'s root and on `usermenu`'s `PartTrigger` — the circle and the pill survived (`usermenu` has a second, pre-existing hit on `PartAvatar`) |
| 7 | `grep -n 'func TestNoHandRolledButtons' conformance_test.go` | one hit |
| 8 | `gotest ./...` | green |
| 9 | `go vet ./... && gofmt -l .` | clean |
| 10 | `grep -rn 'TODO\|FIXME\|Deprecated' --include='*.go' .` | no new hits |

## 11. Stages table

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Bump | `go.mod`, `go.sum` | `widget v0.6.26`, `go mod tidy` clean |
| 2 | Migrate | `actionbutton/css.go`, `scheduleeditor/css.go`, `themetoggle/css.go`, `usermenu/css.go` | 9 `style.Button(` sites; 0 hand-rolled recipes |
| 3 | Leave alone | — | the §5 list is byte-identical to before |
| 4 | Guard | `conformance_test.go` | `TestNoHandRolledButtons` present and green |
| 5 | Verify | — | `gotest`, `go vet`, `gofmt` all green |
