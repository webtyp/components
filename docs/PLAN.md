---
PLAN: "feat!(scheduleeditor): blocks, pattern-apply-to-days, and a bulk day marker"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **No cross-repo gate.** `scheduleeditor` is a component: modules depend on
> it, never the reverse, and this plan imports neither `router`, `orm`,
> `veltylabs/business_calendar` nor `veltylabs/appointment_booking` (§1 already
> stated this; this line makes it explicit at the top so it is never mistaken
> for a dependency). `PatternRow`, `MarkedDay` and `Bounds` (§5) are plain Go
> types this repo defines; the *host* — `app-demo`'s plan — is what translates
> them to and from those two modules' ops. This plan can execute **in
> parallel** with `business_calendar`/`appointment_booking`, not after them.
> Its only real prerequisite is intra-repo: `calendarslider` gains
> multi-select first, as Stage 4a of this same plan. Orchestrator:
> [webtyp/docs/AGENDA_DOMAIN_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_DOMAIN_MASTER_PLAN.md)
> (§3-bis explains the same rule for `appointment_booking`, one level down: a
> generic module never imports a specific one either).
>
> The previous `scheduleeditor` plan is closed and rotated to
> [LAST_PLAN_EXECUTED.md](LAST_PLAN_EXECUTED.md) (published as `v0.6.22`). This
> plan replaces the editor's interaction model; that one only fixed its bugs.

# Plan — `scheduleeditor` v2

Everything is in `scheduleeditor/`. The component stays **pure**: it knows
nothing about `router`, `orm` or `appointment_booking`. The host supplies data
and translates callbacks.

## 1. Why the current editor is the wrong shape

Driven live at `#agenda` (`components v0.6.22`), after the bug-fix round:

- **7 rows × 4 selects, always.** A professional with one pattern
  (Mon/Wed/Fri 08:00–14:00) edits **28 controls** to express one fact. The
  grid's size is fixed by the calendar, not by the schedule.
- **A break is a column.** `Colación desde` / `Colación hasta` can describe
  exactly one interruption, so "morning 09–13, afternoon 15–19" is
  unrepresentable.
- **No way to work an irregular set of days.** The weekly grid is the only
  input; a professional who works "some Saturdays" or "the last working day of
  each month" has nowhere to say it.
- **The exceptions panel is 899px tall** inside a ~544px viewport, dominated by
  three stacked months of `calendarslider`. Measured live.
- **Hours are unbounded.** `hourOptions` hardcodes 06:00–22:00 regardless of
  what the establishment actually allows.

## 2. What it becomes

One view, two mechanisms visible at once (master plan A5) — no mode switch:

```
┌─ Plantilla semanal ────────────────────────────────┐
│  [09:00 ▾] – [13:00 ▾]   [L][M][X][J][V][S][D]  [+]│  ← a pattern row
│  [15:00 ▾] – [19:00 ▾]   [L][M][X][J][V][S][D]  [×]│  ← another
└────────────────────────────────────────────────────┘
┌─ Días marcados ────────────────────────────────────┐
│  [calendario del horizonte, días clicables]        │
│  bloque para los días marcados: [09:00▾]–[13:00▾]  │
└────────────────────────────────────────────────────┘
```

A professional with a fixed pattern uses the top and never touches the bottom.
An irregular one leaves the top empty and marks days below. A mixed one uses
both. **Nothing is hidden and no mode is chosen.**

## 3. Design gate

Breaking change to public API. Per skill **api-design**:

### 3.1 Prior art

- **Cal.com** — a schedule is a list of availability rows, each carrying
  `days: [1,3,5]` + a time range. One row covers several days; a day may appear
  in several rows. There is **no break field** — a break is the gap between two
  rows. Its "Copy times to…" button exists precisely because a per-day grid
  makes the common case expensive.
- **Calendly** — "Weekly hours" as rows with day chips, plus "Date overrides"
  on a calendar. The two live in one screen, exactly as A5 decides.
- **Google Calendar working hours** — per-day ranges, several per day, no break
  concept.

All three converge on *time-range rows tagged with days*, not *days holding a
window*. That inversion is what collapses 28 controls into 2–3.

### 3.2 The novice-name test

- `PatternRow{StartMin, EndMin, Days []int}` — *"this time range, on these
  days."*
- `MarkedDay{Date, StartMin, EndMin}` — *"this date, with these hours."*
- `Bounds{OpenMin, CloseMin}` — *"the earliest and latest the place is open."*
- `OnPatternChange(rows []PatternRow)` — *"the pattern is now this."*
- `OnDaysMarked(dates []string, startMin, endMin int)` — *"mark these days with
  this block."*

### 3.3 Complexity ledger

| Row | Δ |
|---|---|
| Concepts the developer must learn | **−1** — "break" stops being a concept; **+1** — "marked day". Net 0, and the surviving concept is the one that maps to the domain |
| Controls to express one pattern | **28 → 3** (start, end, day chips) |
| Ways to express two sessions a day | **0 → 1** |
| Ways to express an irregular schedule | **0 → 1** |
| Ways to do the same thing | **−1** — `Week []WeeklyRow` is deleted, not kept beside patterns |
| Exported surface | **−1 type** (`WeeklyRow`), **+3 types**, **−1 callback**, **+2 callbacks** |

### 3.4 Where it belongs

`scheduleeditor` owns "edit a professional's availability". Bounding the hour
options to what the establishment allows is *rendering the data it was given* —
`Bounds` arrives as input; the component never fetches it and never validates
against a service. That stays with the host and the domain module.

### 3.5 What it deletes

`WeeklyRow` (and with it `BreakStart`/`BreakFinish`), `OnWeeklyChange`,
`weekHeadKeys` and the five column headers, `PartWeekHead`/`PartWeekHeadCell`,
`PartDay`/`PartDayName`/`PartToggle`, and the unbounded `hourOptions` range.
Nothing is deprecated — the single consumer (`app-demo`) migrates in its plan.

## 4. Use cases

CU-02, CU-08, CU-09, CU-10, CU-11, CU-12, CU-14, CU-15, CU-24 from the master
plan §4. Each gets a test in §9.

## 5. Stage 1 — the new data shape

**File: `scheduleeditor/scheduleeditor.go`.** Delete `WeeklyRow` entirely. Add:

```go
// PatternRow is a time range plus the weekdays it applies to. Several rows may
// cover the same weekday — "morning 09:00–13:00, afternoon 15:00–19:00" is two
// rows sharing a day, and the lunch break is the GAP between them.
//
// There is deliberately no break field: a break that is a field can describe
// exactly one interruption, and a gap describes any number.
type PatternRow struct {
	StartMin, EndMin int   // minutes from midnight
	Days             []int // 0=Sunday … 6=Saturday
}

// MarkedDay is a concrete date the professional works, independent of the
// weekly pattern. It is what lets an irregular schedule exist at all: a
// professional with no PatternRow and a list of MarkedDay is fully expressed.
type MarkedDay struct {
	Date             string // "YYYY-MM-DD"
	StartMin, EndMin int
}

// Bounds is the establishment's opening window — the only hours the editor may
// offer (CU-02). It is INPUT: the component renders within it and never
// fetches or validates it. Zero value (0,0) means unbounded, for a host that
// has no institutional calendar.
type Bounds struct {
	OpenMin, CloseMin int
}
```

`ScheduleEditor` becomes:

```go
type ScheduleEditor struct {
	Element // value embed — NEVER pointer

	// Pattern is the weekly template. Initial state — the host persists on
	// each callback and re-mounts with fresh data.
	Pattern []PatternRow
	// Marked are the concrete dates worked, sorted ascending.
	Marked []MarkedDay
	// Bounds caps every hour control (CU-02). Widening it upstream is what
	// makes CU-03 visible here with no rebuild.
	Bounds Bounds
	// Horizon is how many months the marking calendar shows. 0 → 6.
	Horizon int
	// Holidays and Closures are read-only dates the editor paints as
	// unavailable; picking one is refused by the calendar, not by a message.
	Holidays []string
	Closures []string

	OnPatternChange func(rows []PatternRow)
	OnDaysMarked    func(dates []string, startMin, endMin int)
	OnDaysUnmarked  func(dates []string)
	OnMarkedDayEdit func(day MarkedDay)
}
```

`Exception`, `OnExceptionAdd` and `OnExceptionRemove` **stay** — blocking a day
because you fell ill (CU-13) is still an exception, and it is not the same
gesture as unmarking a working day.

## 6. Stage 2 — bounded hour options (CU-02, CU-03)

Replace the hardcoded 06:00–22:00 range:

```go
// hourOptions returns the <option> set for an hour select, clamped to the
// establishment's opening window. Hours outside it are NOT rendered disabled —
// they are not rendered at all: an option that cannot legally be chosen has no
// reason to exist in the list (CU-02).
//
// Bounds{} (0,0) means the host has no institutional calendar; fall back to
// the full day, 00:00–23:45.
func hourOptions(selected int, b Bounds, step int) []*Element
```

`step` stays 15. When `Bounds` widens upstream and the host re-mounts, the
select simply has more options — that is all CU-03 needs at this layer.

## 7. Stage 3 — the pattern editor (CU-08, CU-09, CU-12)

A pattern row renders as: start select, end select, seven day chips, and a
remove button. Below the rows, one "add row" button.

**Day chips are `<input type="checkbox">` + `<label>`, never divs.** They must be
keyboard-reachable and announce their state. Give each chip a `Key(...)` and
retrieve it with `el.Ref()`; **never compose an id string** — that is the rule
`DEMO_AGENDA_MASTER_PLAN.md` records under "Fix de ids del arnés".

New parts, replacing the deleted grid parts:

```go
	PartPattern     = widget.Part("pattern")
	PartPatternRow  = widget.Part("pattern-row")
	PartDayChips    = widget.Part("day-chips")
	PartDayChip     = widget.Part("day-chip")
	PartRowRemove   = widget.Part("row-remove")
	PartRowAdd      = widget.Part("row-add")
	PartMarker      = widget.Part("marker")
	PartMarkerHours = widget.Part("marker-hours")
```

A row whose `Days` is empty, or whose `StartMin >= EndMin`, or that overlaps
another row on a shared day, carries `widget.Invalid` and logs a dev warning —
**never blocks the edit**, same rule the component already follows.

`OnPatternChange` hands back the **whole** row set, not a delta. A whole-set
replace is what keeps the stored pattern and the rendered pattern from drifting,
and it matches `SaveDayBlocks`' whole-day replace upstream.

## 8. Stage 4 — the day marker (CU-10, CU-11, CU-15, CU-24)

A `calendarslider` over `Horizon` months (default 6) where **every** day is
selectable, plus one start/end pair for the common window (A4).

- Clicking an unmarked day marks it with the common window → `OnDaysMarked`.
- Clicking a marked day unmarks it → `OnDaysUnmarked`.
- A marked day shows its own hours when they differ from the common window;
  editing them fires `OnMarkedDayEdit` (CU-11).
- Holidays and closures render unavailable and are not clickable.

**CU-24 — the size problem.** The old panel was 899px because it stacked three
months. `calendarslider` already has a collapsed mode
(`calendarslider__collapsed`). Render the marker **collapsed by default**,
showing the current month, and let the strip scroll horizontally. Read
`calendarslider`'s API before wiring: `NumMonths`, `Occupation`, `Selected`,
`OnSelect` and the collapse toggle already exist — do not add a second
calendar and do not fork it.

### 8.1 `calendarslider` must gain multi-selection first — verified

`calendarslider` is **single-select today**. Confirmed in
`calendarslider/calendarslider.go`: `Selected *SignalString` (line 123), and
selection is derived as `isSel := DeriveBool(func() bool { return
c.Selected.Get() == dateStr })` (line 509). One date at a time.

Marking days is inherently multi-select, so **the capability is added to
`calendarslider`, in this repo, as Stage 4a** — never worked around inside
`scheduleeditor`. Working around a missing capability in the consumer is the
fork this ecosystem forbids, and a second calendar would be worse.

```go
// SelectedMany holds every selected date key when the calendar is in
// multi-select mode. nil (the default) keeps the existing single-select
// behaviour through Selected, so every current consumer is untouched.
SelectedMany *SignalStrings // nil => single-select via Selected

// OnToggle fires in multi-select mode with the date and its new state.
OnToggle func(date string, selected bool)
```

Both modes coexist: `SelectedMany == nil` is exactly today's behaviour, so
`targethour`, `reservation` and the exceptions panel keep working unchanged.
Check whether `dom` exposes a `SignalStrings`; if it does not, use a
`*SignalString` carrying a delimited key set rather than adding a signal type —
and say which you chose in the PR.

`calendarslider` tests must cover: single-select unchanged, multi-select toggles
on and off, and `Occupation`/`Holidays` still paint correctly in both modes.

## 9. Stage 5 — CSS

**File: `scheduleeditor/css.go`.** Delete the rules for the removed parts. Add
`PartPattern` (`Stack`), `PartPatternRow` (`Row` + `ControlBox` + `Round`),
`PartDayChips` (`Row`), `PartDayChip` (`ControlBox` + `Round` + `Interactive`),
`PartMarker` (`Stack` + `As(Panel)`), `PartMarkerHours` (`Row`).

**One flow primitive per `Part`.** `Row`, `Stack`, `Grid`, `FixedGrid`,
`Center`, `Split` and `ScrollRow` all assign `rule.flowType` and the last one
silently wins — that defect shipped twice in the previous round. `Row(gap)`
already emits `align-items: center`; it never needs `Center()`.

Every interactive element ≥ 44×44 via `ControlBox()`. A day chip is a tap
target; verify with `browser_audit_mobile`.

## 10. Stage 6 — tests

`scheduleeditor_test.go` and `scheduleeditor_ui_wasm_test.go` construct
`WeeklyRow` and `OnWeeklyChange` throughout. Find every one:
`grep -rn "WeeklyRow\|OnWeeklyChange\|BreakStart\|BreakFinish" scheduleeditor/`.

| Test | CU |
|---|---|
| `TestOnePatternRowCoversSeveralWeekdays` | CU-08 |
| `TestTwoRowsShareADayAndLeaveAGap` | CU-09 |
| `TestEmptyPatternWithMarkedDaysIsValid` | CU-10 |
| `TestMarkingDaysUsesTheCommonWindow` | CU-11 |
| `TestMarkedDayCanDivergeFromTheCommonWindow` | CU-11 |
| `TestPatternAndMarkedDaysRenderTogether` | CU-12 |
| `TestUnmarkingADayFiresOnDaysUnmarked` | CU-15 |
| `TestHourOptionsAreClampedToBounds` | **CU-02** |
| `TestWiderBoundsOfferMoreOptions` | **CU-03** |
| `TestZeroBoundsFallBackToFullDay` | — |
| `TestHolidayIsNotSelectableInTheMarker` | CU-14 |
| `TestOverlappingRowsAreMarkedInvalidButNotBlocked` | — |
| `TestRevealedStatesAreWrittenByTheMarkup` | keep from the previous plan |

Verified facts to reuse, do not re-derive: `fmt.KeyValue` exposes `Key`/`Value`
as **fields**; `testEditor()` and `emptyCtx{}` already exist in
`scheduleeditor_test.go`, which carries `//go:build !wasm` and imports
`strings`, `testing`, `webtyp.com/date`; attributes serialize as `key='value'`
with **single** quotes, so counting a bare word double-counts.

## 11. Stage 7 — README

Rewrite the data-shape, usage and behaviour sections. New translation keys to
list: `Add row`, `Remove row`, `Marked days`, `Weekly pattern`, `Hours for
marked days`, plus the seven weekday **short** names for the chips.

## 12. Constraints

- **No stdlib in WASM code**: `webtyp.com/fmt`, never `strconv`/`strings`.
  `_test.go` files already import `strings`; production code must not.
- **`dom.Element` embedded by VALUE.**
- **`css.go` carries `//go:build !wasm`.** Never `ssr.go`, never `front.go`.
- **Never compose an id string in `Render()`** — `Key(...)` + `el.Ref()`,
  `data-*` as breadcrumb.
- **No `map`** in code reaching the WASM binary.
- **The component stays pure** — no `router`, no `orm`, no domain import.
- **Do not fork `calendarslider`.** Missing capability is fixed there (§8).
- **No `TODO`, nothing deprecated.** `WeeklyRow` is deleted. Before closing:
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' scheduleeditor/`
- `gotest`, never `go test`.

## 13. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./scheduleeditor/` | green, every test in §10 |
| 2 | `gotest ./...` | green — `conformance_test.go` included |
| 3 | `grep -rn "WeeklyRow\|OnWeeklyChange\|BreakStart\|BreakFinish" scheduleeditor/` | **empty** |
| 4 | `grep -rn "360\|1320" scheduleeditor/scheduleeditor.go` | **empty** — the hardcoded range is gone |
| 5 | `grep -c "style.Center()" scheduleeditor/css.go` | **0** |
| 6 | `GOOS=js GOARCH=wasm go build ./...` | compiles |
| 7 | `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' scheduleeditor/` | no new hit |

## 14. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | New data shape | `scheduleeditor.go` | `WeeklyRow` gone; `PatternRow`/`MarkedDay`/`Bounds` in |
| 2 | Bounded hours | `scheduleeditor.go` | CU-02 and CU-03 green |
| 3 | Pattern editor | `scheduleeditor.go` | CU-08, CU-09, CU-12 green |
| 4a | `calendarslider` multi-select | `calendarslider/` | `SelectedMany`/`OnToggle`; single-select consumers untouched |
| 4b | Day marker | `scheduleeditor.go` | CU-10, CU-11, CU-15 green |
| 5 | CSS | `css.go` | one flow primitive per Part; 44×44 targets |
| 6 | Tests | both `_test.go` | §10 complete |
| 7 | README | `README.md` | shapes, keys, behaviour current |
