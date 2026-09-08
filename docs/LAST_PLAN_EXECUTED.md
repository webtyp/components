---
PLAN: "fix(calendarslider): drop author-invented global element ids"
TAG: v0.6.18
EXECUTOR: local
REVIEWER: none
---

> Executed LOCALLY (not dispatched). Wave 1 of 2 — wave 2 is `webtyp/dom`
> making an author-set id inside a component's Render a hard panic.

## Execution notes — two deviations from the written spec below

1. **A third `ID()` the initial grep missed:** `buildCollapsed`'s hidden
   checkbox used `ID(c.uid+"-collapsed-toggle")` + a `for=` label. Fixed by
   **nesting the checkbox inside the `<label>`** (implicit association — no
   `for`, no id). `c.uid`, `nextCalendarSliderID`, `calendarSliderSeq`,
   `suffixCollapsedToggle` deleted as now-dead.
2. **Month card resolution:** the spec had `slideToMonth` call
   `Get(monthEl.GetID())`. Minting the id at build time broke
   `TestCalendarSlider_RenderIdempotent` (counter-based id ≠ deterministic
   across renders — the same defect flagged in the closed dom PR #23). Instead
   the card carries **no id**; `c.months` stores each month's `‹` **button**
   (already dom-id'd in WASM because it has a click handler, and never in SSR —
   the id path is observer-gated), and `slideToMonth` scrolls that button into
   view, which snaps the full-width strip to its card. Same pattern
   `usermenu` uses (`Get(menu.GetID())` on an element that carries an event).

`data-month` / `data-date` breadcrumbs replace the `#cs-m-` / `#cs-d-`
selectors in tests. New WASM regression: `TestTwoInstancesShareAPage`.

---

# PLAN — `calendarslider` stops inventing global element ids

This is **wave 1 of 2**. Wave 2 (`webtyp/dom`) makes an author-set id inside a
component's `Render()` a hard panic. This plan must land and publish **first**,
so that `dom`'s new rule finds nothing to panic on.

## Problem

`app-demo` mounts TWO `CalendarSlider` instances in one render (the `reservation`
filter and `agenda`/`ScheduleEditor`). Both emit the same hardcoded global ids,
so `dom.claimID` panics:

```
dom: id cs-m-2026-09 was written twice in one render, by <div> and <div>
```

The ids are invented by the component from data it does not uniquely own
(`"cs-m-" + monthKey`), then used as a cross-instance lookup handle.

## Decision (closed with the framework owner — not negotiable)

**A component never creates a global element id.** `dom` owns ids. The component
identifies its own nodes with `Key()` and, when it needs the live DOM node,
resolves it through the element it already holds:

```go
ref, ok := Get(el.GetID())   // GetID() mints and caches a dom id on demand
```

`(*Element).GetID()` already auto-generates and caches an id when the element has
none — see
[element.go:249](https://github.com/webtyp/dom/blob/main/element.go#L249). No new
`dom` API is required by this plan, and `dom` does not need to change for this
plan to go green.

A previously dispatched plan proposed instance-prefixing the ids
(`cs-m-<instance>-2026-09`). **That approach is rejected**: it keeps the component
in the business of minting global ids, which is the thing being removed. Do not
reintroduce it.

## Anti-footguns

- This repo compiles to WASM. **No `map`** — the new lookup is a slice with a
  linear scan (a strip holds a handful of months). Do not "optimize" it into a
  `map[string]*Element`; the map runtime is binary budget this ecosystem does not
  spend.
- **No standard library** in this package: use `webtyp/fmt`, never `strings` /
  `strconv` / `errors`.
- Do **not** touch `webtyp/dom` from this plan. Wave 2 owns that repo.
- Do **not** remove `Key()` anywhere. `Key` is the author's identity contract and
  `BindChildren` reconciles on it.

## Stage 1 — `calendarslider/calendarslider.go`: hold the month elements

Add, next to the other package types:

```go
// monthRef pairs a month key with the element built for it, so slideToMonth can
// reach the live node without inventing a global id. A slice, not a map: the
// strip holds a handful of months and this package compiles to WASM, where the
// map runtime is budget this ecosystem does not spend.
type monthRef struct {
	key string
	el  *Element
}
```

Add the field to `CalendarSlider`:

```go
	// months records the element built for each month key in the current
	// render, so the ‹ › arrows can scroll to a sibling card. Rebuilt from
	// scratch on every Render (see Render).
	months []monthRef
```

## Stage 2 — `buildMonth`: drop the id, record the element

In `func (c *CalendarSlider) buildMonth(...)` (currently
`calendarslider.go:337`):

- **Delete** the `ID("cs-m-" + key)` call (`calendarslider.go:341`).
- Keep `Key(key)`.
- Add `Attr("data-month", key)` in its place — the same test/CSS-addressable
  breadcrumb the day cells already use with `data-date`, and **not** an id.
- Record the element before returning it:

```go
	c.months = append(c.months, monthRef{key: key, el: monthEl})
	return monthEl
```

So the head of `buildMonth` becomes:

```go
	key := date.MonthKey(year, month)
	monthEl := Div().Set(clsMonth.AsAttr()).
		Key(key).
		Attr("data-month", key)
```

## Stage 3 — `Render`: reset the slice before rebuilding

`Render()` rebuilds every month card on each call, so the slice must not grow
without bound. Immediately before the loop that calls `buildMonth` (around
`calendarslider.go:231`, where `keys[i] = date.MonthKey(y, m)` is filled), add:

```go
	c.months = c.months[:0] // rebuilt below; reuse the backing array
```

## Stage 4 — `slideToMonth` becomes a method

Replace the package-level `func slideToMonth(key string, instant bool)`
(`calendarslider.go:403`) with a method. It was package-level only because it had
no way to reach the instance — which is exactly why it reached for a global id.

```go
// slideToMonth jumps the scroll-snap strip to the month card carrying the given
// key. instant selects ScrollIntoViewInstant over the normal smooth
// ScrollIntoView — reserved for the two wrap edges (first month's ‹, last
// month's ›), where a smooth scroll would visibly travel across every month in
// between in the wrong apparent direction. Every adjacent-month navigation keeps
// calling this with instant=false.
//
// The card is resolved through the element this instance built, never through a
// global id: two calendars on one page each scroll their own strip.
func (c *CalendarSlider) slideToMonth(key string, instant bool) {
	var el *Element
	for i := range c.months {
		if c.months[i].key == key {
			el = c.months[i].el
			break
		}
	}
	if el == nil {
		return
	}
	ref, ok := Get(el.GetID())
	if !ok {
		return
	}
	if instant {
		ref.ScrollIntoViewInstant()
		return
	}
	ref.ScrollIntoView()
}
```

Update both call sites (`calendarslider.go:380` and `:387`):

```go
	prev.On("click", func(Event) { c.slideToMonth(prevKey, prevWraps) })
	next.On("click", func(Event) { c.slideToMonth(nextKey, nextWraps) })
```

## Stage 5 — the arrow breadcrumbs stop carrying an id

`calendarslider.go:378` and `:385` set `Attr("data-target", "cs-m-"+prevKey)` /
`"cs-m-"+nextKey`. Nothing reads these at runtime; they exist as a breadcrumb and
are asserted by one SSR test. Drop the `cs-m-` prefix so no code anywhere
constructs an id-shaped string:

```go
		Attr("data-target", prevKey).
```
```go
		Attr("data-target", nextKey).
```

## Stage 6 — day cells drop their id

At `calendarslider.go:487-489`, **delete** `ID("cs-d-"+dateStr)`. Keep `Key(dateStr)`
and the already-present `Attr("data-date", dateStr)` — that attribute is what
tests address.

## Stage 7 — tests

`calendarslider_test.go` (SSR, `!wasm`):

- Line 39-40: `data-target='cs-m-2026-07'` → `data-target='2026-07'`.

`calendarslider_wasm_test.go` (`wasm`): every selector built on the deleted ids
moves to the data attributes.

- `#cs-d-2026-08-11` → `[data-date='2026-08-11']` (lines 51, 62, 154, 165).
- `#cs-d-2026-08-01` → `[data-date='2026-08-01']` (line 70).
- `#cs-m-2026-08 .calendarslider__month-name` →
  `[data-month='2026-08'] .calendarslider__month-name` (line 89).
- `#cs-m-2026-10 .calendarslider__month-name` →
  `[data-month='2026-10'] .calendarslider__month-name` (line 92).

New test in `calendarslider_wasm_test.go` — the regression this whole wave exists
for:

```go
// TestTwoInstances_NoIdCollision mounts two calendars in ONE render, the shape
// app-demo uses (a reservation filter and the agenda editor). Before this, both
// emitted id="cs-m-<month>" and dom.claimID panicked. Each must now render and
// scroll its own strip.
func TestTwoInstances_NoIdCollision(t *testing.T) { ... }
```

It must assert:
1. Rendering both in one pass does not panic.
2. `document.querySelectorAll("[data-month='2026-08']")` returns **2** nodes,
   and their `id` attributes are **different and non-empty**.
3. Clicking the ‹ arrow inside instance A scrolls A's strip and leaves B's
   `scrollLeft` untouched (stub `Element.prototype.scrollIntoView` the way
   `calendarslider_wasm_test.go:181` already does, and record which node it was
   called on).

## Grep-verifiable acceptance criteria

- `grep -rn 'cs-m-\|cs-d-' calendarslider/` → **empty** (production code and tests).
- `grep -rn 'ID("' calendarslider/*.go | grep -v _test` → **empty**.
- `grep -rn 'map\[' calendarslider/*.go` → **empty**.
- `grep -n 'func slideToMonth' calendarslider/calendarslider.go` → **empty**
  (it is a method now).
- `gotest ./...` green, `wasm ✅` included.
- `gofmt -l .` → **empty**.

## Out of scope

- `webtyp/dom` — wave 2, its own plan and repo.
- Instance-prefixed ids (`cs-m-<instance>-…`) — rejected, see Decision.
- Any other component. A sweep of this repo shows `calendarslider` is the **only**
  component that sets an element id inside `Render()`; the two other hits
  (`selectsearch/web/client.go`, `calendarslider/web/client.go`) are demo `main`s
  where a root-level `ID("app-result")` is legitimate and stays.

## Stages

| # | File | Change |
|---|---|---|
| 1 | `calendarslider/calendarslider.go` | `monthRef` type + `months` field |
| 2 | `calendarslider/calendarslider.go` | `buildMonth`: drop `ID()`, add `data-month`, record element |
| 3 | `calendarslider/calendarslider.go` | `Render`: reset `c.months` |
| 4 | `calendarslider/calendarslider.go` | `slideToMonth` → method resolving via `GetID()` |
| 5 | `calendarslider/calendarslider.go` | `data-target` drops the `cs-m-` prefix |
| 6 | `calendarslider/calendarslider.go` | day cells drop `ID("cs-d-…")` |
| 7 | `calendarslider/*_test.go` | selectors → `data-date` / `data-month`; new two-instance test |
