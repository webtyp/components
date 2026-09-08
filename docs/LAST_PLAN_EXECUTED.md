---
PLAN: "fix(scheduleeditor): the day is the index, the reveal is real, the grid is legible"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Depends on a GATE.** `webtyp.com/widget` must first ship
> `Kind.Allows(Form, Open) == true` (see
> [widget/docs/PLAN.md](https://github.com/webtyp/widget/blob/main/docs/PLAN.md)).
> Bump `webtyp.com/widget` in `go.mod` to that tag **before** Stage 2 — without
> it, `TestKindAllowsEveryState` in this repo fails. Orchestrator:
> [webtyp/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md).

# Plan — `scheduleeditor` correctness + legibility

Everything here is in **one package**: `scheduleeditor/`. Files touched:
`scheduleeditor.go`, `css.go`, `README.md`, `scheduleeditor_test.go`,
`scheduleeditor_ui_wasm_test.go`.

---

## 1. What is broken (observed in the running demo, not inferred)

`ScheduleEditor` is mounted at `https://localhost:8080/#agenda` by
`app-demo/modules/agenda`. Driven live at `components v0.6.20`:

1. **The day label is taken from data the host can get wrong.** The rows render
   `Domingo · Lunes · Domingo · Miércoles · Domingo · Viernes · Domingo`. The
   host built `make([]WeeklyRow, 7)` and filled only the days its op returned;
   every unfilled row kept `DayOfWeek == 0`, and `buildWeekRow` renders
   `dayLabel(r.DayOfWeek)`. Worse, `weeklyChange` hands that same lying row to
   `OnWeeklyChange` and the host persists it — **enabling Tuesday wrote
   Sunday**, reproduced live.
2. **`data-open` is written and never styled.** `buildExceptionForm` binds
   `data-open` on `PartExcForm`, and `css.go` has no reveal rule for it, so the
   add-form is permanently visible — measured `display: flex` with no day
   picked. Pressing "Agregar" then hits `addException`'s `if e.sel.Get() == ""
   { return }`: a **silent no-op**.
3. **`PartExcHours` is not in the sheet at all.** No `Part(PartExcHours, …)`
   call exists, so its `data-open` is inert too and the hour selects show for
   `HOLIDAY`, which has no hour window.
4. **The weekly grid has no column headers** — four bare `<select>` per row —
   and `style.Grow()` on `PartDayName` stretches the label to 584px of a 926px
   row, shoving the controls against the far edge.
5. **An inactive day looks like a configured one**: `06:00 06:00 06:00 06:00`,
   selects fully enabled.
6. **`lang.Translate("Special", "hours")` produces "Especial horas"** —
   word-by-word translation cannot order a Spanish adjective phrase.
7. **13 tap targets under 44×44px** (`browser_audit_mobile`): the 7 day toggles
   and 3 type radios render at 13×13. On a 394px viewport a radio and its label
   split across a line break.
8. **An empty exception list renders nothing** — no empty state.

---

## 2. Design gate

This plan changes public API. Per skill **api-design**, the five answers:

### 2.1 Prior art

The question is: *how does a repeating-row editor tell the caller which row
changed, without letting the caller misstate it?*

- **React Hook Form `useFieldArray`** (JS). The array index is the identity;
  `fields[index]` and the change callback both carry `index`. The row object
  never stores its own position — storing it is a documented anti-pattern
  because the two desynchronise on insert/remove.
- **Angular `FormArray`**. `at(index)` is the only accessor; `valueChanges`
  emits the whole array. Position is structural, never a field.
- **Rails `fields_for` / `accepts_nested_attributes_for`** (Ruby). The nested
  record *does* carry an `id`, precisely because rows are unordered database
  rows with no natural position. The position-as-identity model is chosen only
  when the collection is fixed-length and positionally meaningful.

A week is exactly that third case inverted: **fixed length 7, and the position
is the meaning.** Sunday is not "the row whose `day_of_week` happens to be 0",
it is "the first row". So this ecosystem follows family one — index-as-identity,
like `useFieldArray` — not because it is novel but because a fixed-length
positional collection cannot benefit from a redundant position field. The field
can only ever agree with the index (noise) or disagree with it (the bug).

### 2.2 The novice-name test

Read aloud, with no context:

- `OnWeeklyChange func(dayOfWeek int, row WeeklyRow)` — *"on weekly change, give
  me the day of week and the row."* `dayOfWeek` is the word
  `appointment_booking` already uses in `work_calendar_weekly.day_of_week` and
  the word `date.WeekdayName(int)` takes. No lookup needed.
- `PartWeekHead` / `PartWeekHeadCell` — *"the week's header, and a cell of it."*
  Sits beside the existing `PartWeek` / `PartWeekRow` with no new vocabulary.
- `WeeklyRow{Active, WorkStart, WorkFinish, BreakStart, BreakFinish}` — every
  remaining field answers *"what does this day look like"*. None answers *"which
  day is this"*, which is now unaskable.

### 2.3 The complexity ledger

| Row | Δ |
|---|---|
| Concepts the developer must learn | **−1** — "the slice index and the `DayOfWeek` field must agree" stops being a rule anyone must know |
| Files they must touch to do X | **0** |
| Lines at the call site | **+1** in `OnWeeklyChange` (one extra parameter), **−7** in the host, which no longer assigns `DayOfWeek` per row |
| Ways to do the same thing | **−1** — today `Week[2].DayOfWeek` may be `2` or `0` and both compile; after, there is one |
| Exported surface | **−1 field**, **+2 `widget.Part` constants**, **+1 callback parameter** |

The last row is negative. The surface row is honest: two new parts are added for
the header, and they are the minimum a labelled grid needs.

### 2.4 Where it belongs

`scheduleeditor` owns "edit a weekly schedule template plus per-date
exceptions". The day-of-week↔index invariant is entirely inside that concern —
it is not the host's business to maintain, which is precisely why the host got
it wrong. Enforcing it here is SRP, not scope creep. Nothing moves to another
package; no new package appears.

### 2.5 What it deletes

- `WeeklyRow.DayOfWeek` — the field, its zero value, and the class of bug it
  created.
- `dayLabel(dow int)`'s reliance on caller-supplied data (it now takes the
  index).
- Two dictionary keys downstream (`"Special"`, `"hours"`) — deleted in Stage C,
  the `app-demo` plan.
- Nothing is deprecated. `WeeklyRow.DayOfWeek` is removed in this change, and
  the single consumer (`app-demo/modules/agenda`) is migrated in Stage C. There
  are **zero** external users: verified with
  `grep -rln "scheduleeditor" --include=*.go` across the workspace — the hits
  are this package, `components/conformance_test.go`, and `app-demo`.

---

## 3. Stage 1 — the day is the index

**File: `scheduleeditor/scheduleeditor.go`**

### 3.1 The type

```go
// WeeklyRow is one day of the weekly template. Times are minutes from
// midnight (0..1439); a break of 0/0 means no break.
//
// The row does NOT carry its day: its position in ScheduleEditor.Week is the
// day (index 0 = Sunday … 6 = Saturday). A field would be free to disagree
// with the position, and did — a host that left it at its zero value made
// every unconfigured day claim to be Sunday, and persisted it.
type WeeklyRow struct {
	Active                  bool
	WorkStart, WorkFinish   int
	BreakStart, BreakFinish int
}
```

Delete the `DayOfWeek` field. Delete nothing else from the struct.

### 3.2 The callback

On `ScheduleEditor`:

```go
	// OnWeeklyChange fires on every edit of the weekly template (toggle,
	// entry, exit, break). dayOfWeek is 0=Sunday … 6=Saturday, taken from the
	// row's position in Week — never from the row.
	OnWeeklyChange    func(dayOfWeek int, row WeeklyRow)
	OnExceptionAdd    func(Exception)
	OnExceptionRemove func(id string)
```

### 3.3 The funnel

```go
func (e *ScheduleEditor) weeklyChange(i int, mutate func(*WeeklyRow)) {
	if i < 0 || i >= len(e.Week) {
		Log("scheduleeditor: weekly row index out of range", i)
		return
	}
	r := e.Week[i]
	mutate(&r)
	if e.OnWeeklyChange != nil {
		e.OnWeeklyChange(i, r)
	}
}
```

### 3.4 Label and validity read the index

```go
// dayLabel is the day name for a row's POSITION in Week. The component is a
// library: it renders webtyp/date's canonical English name (Sunday..Saturday)
// through lang.Translate — the app registers the dictionary.
func dayLabel(index int) string {
	return lang.Translate(date.WeekdayName(index)).String()
}
```

In `buildWeekRow(i int, r WeeklyRow)`:
- `Span().Set(clsDayName.AsAttr()).Text(dayLabel(r.DayOfWeek))` →
  `…Text(dayLabel(i))`.
- In the invalid-row `Log`, `date.WeekdayName(r.DayOfWeek)` →
  `date.WeekdayName(i)`.
- Add `Attr("data-day", fmt.Convert(i).String())` on the row, as a breadcrumb
  for tests and for the acceptance checks. It is a `data-*` attribute, **not**
  an id — never compose an id string in `Render()`.

### 3.5 `Week` must be exactly 7 rows — loudly

`buildWeek` currently iterates whatever it is given. A host that passes 5 rows
would silently render a 5-day week. Make it loud and correct:

```go
// weekDays is the fixed length of the weekly template: Sunday..Saturday.
const weekDays = 7

func (e *ScheduleEditor) buildWeek() *Element {
	week := Div().Set(clsWeek.AsAttr()).Attr("role", "grid")
	week.Child(e.buildWeekHead())
	if len(e.Week) != weekDays {
		Log("scheduleeditor: Week must hold exactly 7 rows (Sunday..Saturday), got", len(e.Week))
	}
	for i := 0; i < weekDays && i < len(e.Week); i++ {
		week.Child(e.buildWeekRow(i, e.Week[i]))
	}
	return week
}
```

`Log` is `dom`'s development diagnostic — loud in dev, gone in a release build.
It reports and continues; it does not panic and it does not pad the slice with
guesses. A missing input is never guessed (skill **api-design**, zero technical
debt).

---

## 4. Stage 2 — the reveal actually reveals

**Do not start Stage 2 until `go.mod` points at the `widget` tag from the gate
plan.** Verify first:

```
grep -n "webtyp.com/widget" go.mod          # must be the new tag
```

### 4.1 `css.go` — add the two missing rules

**File: `scheduleeditor/css.go`.** In `sheet()`, change the `PartExcForm` rule
and add a `PartExcHours` rule:

```go
		Part(PartExcForm,
			style.RevealedBy(widget.Open),
			style.Stack(style.Space2),
			style.As(style.Inset),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
		).
		Part(PartExcHours,
			style.RevealedBy(widget.Open),
			style.Row(style.Space2),
		).
```

`widget` is already imported in this file (it is used by
`When(widget.Invalid, …)`). `PartExcHours` is already declared in
`scheduleeditor.go` and already carried on the markup — it was simply never
given a rule.

### 4.2 The test that closes the loop

`widget/docs/DESIGN.md` §17 states why this test must exist: the Go half writes
attributes, the CSS half writes selectors, they live behind different build
tags, and **nothing** — not the compiler, not `Validate()` — connects them.
`Sheet.StateAttrs()` returns the list to assert against. That test is missing
today, which is how this bug shipped.

**File: `scheduleeditor/scheduleeditor_test.go`** — add:

```go
// Every state the stylesheet reveals on must actually be written by the
// markup. The two halves live behind different build tags and nothing checks
// them: this is the loop widget/docs/DESIGN.md §17 says the consumer closes.
func TestRevealedStatesAreWrittenByTheMarkup(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	for _, kv := range e.sheet().StateAttrs() {
		if !strings.Contains(html, kv.Key) {
			t.Errorf("stylesheet reveals on %q but no element writes it:\n%s", kv.Key, html)
		}
	}
}
```

Three facts this test depends on, all verified — do not re-derive them:

- **`fmt.KeyValue` exposes `Key` and `Value` as struct FIELDS**, not methods
  (`webtyp.com/fmt/parse.go`). Write `kv.Key`, never `kv.Key()`.
- **`testEditor()` and `emptyCtx{}` already exist** at the top of
  `scheduleeditor_test.go`. Use them; do not add a second fixture.
- **`Sheet.sheet()` is unexported and this test is in-package** — that is why it
  can reach `e.sheet()`. Keep the test in `scheduleeditor_test.go`.

---

## 5. Stage 3 — a legible weekly grid

### 5.1 New parts

**File: `scheduleeditor/scheduleeditor.go`** — add to the `Part` const block and
the `cls…` var block, following the existing style exactly:

```go
	PartWeekHead     = widget.Part("week-head")
	PartWeekHeadCell = widget.Part("week-head-cell")
```

```go
	clsWeekHead     = NameScheduleEditor.Class(PartWeekHead)
	clsWeekHeadCell = NameScheduleEditor.Class(PartWeekHeadCell)
```

### 5.2 The header row

```go
// weekHeadKeys are the five column labels, in render order. English keys —
// the app's dictionary translates them (see README, "Translation keys").
var weekHeadKeys = []string{"Day", "Work start", "Work end", "Break start", "Break end"}

// buildWeekHead is the label row of the weekly grid. Without it the four
// selects are four unlabelled dropdowns and nothing says which is which.
func (e *ScheduleEditor) buildWeekHead() *Element {
	head := Div().Set(clsWeekHead.AsAttr()).Attr("role", "row")
	for _, k := range weekHeadKeys {
		head.Child(Span().Set(clsWeekHeadCell.AsAttr()).
			Attr("role", "columnheader").
			Text(lang.Translate(k).String()))
	}
	return head
}
```

### 5.3 Inactive days dim and disable

In `buildWeekRow`, after the four `timePick` children are appended, the selects
of an inactive day must be non-interactive — a disabled control cannot be
half-edited into a meaningless state, and it is what makes `06:00 06:00 06:00
06:00` read as "not configured" instead of "configured at 06:00".

`timePick` gains a `disabled` parameter:

```go
func timePick(part widget.Part, name string, val int, disabled bool, onChange func(int)) *Element {
	sel := NewElement("select").
		Set(NameScheduleEditor.Class(part).AsAttr()).
		Attr("name", name)
	if disabled {
		sel.Attr("disabled", "disabled")
	}
	// … options and the change listener, unchanged …
}
```

Every `timePick(PartTime, "work-start", r.WorkStart, …)` call in `buildWeekRow`
passes `!r.Active`. `boundTimePick` (the exception form's two selects) is a
different function and is **not** changed.

The row already carries `Attr("data-active", mapBool(r.Active))`. Keep it — the
CSS dims on it.

### 5.4 `css.go` — the grid

Replace the `PartWeek`, `PartWeekRow`, `PartDay`, `PartDayName` and `PartTime`
rules, and add the two header rules. The rest of `sheet()` is untouched except
for §4.1.

```go
		Part(PartWeek,
			style.Stack(style.Space1),
		).
		Part(PartWeekHead,
			style.FixedGrid(5, style.Space2),
			style.PadInline(style.Space2),
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
			style.As(style.Subtle),
		).
		Part(PartWeekHeadCell,
			style.KeepSize(),
		).
		Part(PartWeekRow,
			style.FixedGrid(5, style.Space2),
			style.ControlBox(),
			style.Round(style.RadiusMd),
			style.Anchor(),
		).
		When(widget.Invalid, PartWeekRow,
			style.As(style.DangerWash),
		).
		Part(PartDay,
			style.Row(style.Space2),
		).
		Part(PartDayName,
			style.FontWeight(style.WeightBold),
		).
		Part(PartToggle,
			style.ControlBox(),
			style.KeepSize(),
		).
		Part(PartTime,
			style.ControlBox(),
			style.Round(style.RadiusSm),
			style.As(style.Inset),
		).
```

**Never list two flow primitives in one `Part`.** `Row`, `Stack`, `Grid`,
`FixedGrid`, `Center`, `Split` and `ScrollRow` all assign the same
`rule.flowType` slot (`widget/style/flow.go`), so the last one silently wins and
nothing warns. In particular `Center()` is **not** cross-axis centering — it is
"a centered column with an optional maximum size (defaults to `Readable`)",
emitting `margin-inline: auto; max-width: …; width: 100%`. `Row(gap)` already
emits `align-items: center` (`widget/style/emit_flowdecls.go:21-22`), so a row
never needs it.

The three deletions that matter, each deliberate:

- **`style.Grow()` is gone from `PartDayName`.** It was the 584px dead gap: a
  grid column already sizes the cell, and `Grow()` inside it stretched the label
  to eat the row.
- **`style.KeepSize()` is gone from `PartTime` and `PartDay`.** In a
  `FixedGrid` the column decides the width; `KeepSize()` fought it and produced
  the 61.6px selects crushed against the right edge.
- **`style.KeepSize()` stays on `PartToggle` and `PartWeekHeadCell`** — the
  checkbox must not stretch, and a header cell must not shrink below its label.

`style.FixedGrid(cols int, gap Space)`, `style.PadInline`, `style.Center`,
`style.ControlBox` and `style.Subtle` all already exist in
`webtyp.com/widget/style`; `statgrid/css.go` in this repo is a worked example of
a grid root.

### 5.5 Tap targets ≥ 44×44

`style.ControlBox()` on `PartToggle` and on `PartExcType` gives the checkbox and
the radios a real control box instead of the browser's 13×13 default. Add it to
`PartExcType`'s existing rule:

```go
		Part(PartExcType,
			style.Row(style.Space2),
			style.ControlBox(),
		).
```

After Stage 5's verification, if `browser_audit_mobile` still reports any
element of this component under 44×44, raise its box with
`style.IconBox(style.IconMd)` — check the `IconSize` constants actually exported
by `widget/style` before using a name. Do **not** invent a CSS literal; every
value comes from a token.

### 5.6 The empty exception list

In `buildExceptionList`, when there is nothing to show, say so instead of
rendering an empty `<ul>`:

```go
	if len(items) == 0 && len(e.Holidays) == 0 {
		list.Child(Li().Set(clsExcItem.AsAttr()).
			Text(lang.Translate("No exceptions").String()))
		return list
	}
```

Place it after `items := sortedExceptions(e.Exceptions)` and before the loop.
Read the existing function first — if holidays are rendered in the same loop,
the guard must account for both, exactly as written above.

---

## 6. Stage 4 — "Horario especial", one key

**File: `scheduleeditor/scheduleeditor.go`**

```go
func exceptionTypes() []fmt.KeyValue {
	return []fmt.KeyValue{
		{Key: ExcHoliday, Value: lang.Translate("Closed").String()},
		{Key: ExcSpecialHours, Value: lang.Translate("Special hours").String()},
		{Key: ExcBlocked, Value: lang.Translate("Blocked").String()},
	}
}
```

Why this works: `lang.lookupWord` binary-searches the **whole** argument string
against the dictionary's EN column, case-insensitively
(`webtyp.com/fmt/lang/dictionary.go`). A multi-word key is one lookup.
`lang.Translate("Special", "hours")` was two lookups joined by a space, which
cannot order a Spanish adjective phrase — "Especial horas" instead of "Horario
especial".

**Anti-footgun.** Do not "fix" the other `lang.Translate` calls in this package
by merging their arguments. Every other call in `scheduleeditor.go` passes a
single key already (`"Closed"`, `"Blocked"`, `"Add"`, `"Remove"`, `"Type"`,
`"Date"`, `"Notes"`), and `date.WeekdayName` returns a single word by
construction. Only `"Special hours"` was split.

---

## 7. Stage 5 — README and tests

### 7.1 `scheduleeditor/README.md`

Update, in place:

- The `WeeklyRow` snippet — remove `DayOfWeek`, and state that the slice index
  is the day, `Week` is exactly 7 rows, Sunday first.
- The `OnWeeklyChange` usage snippet — new two-parameter signature.
- **"Translation keys"** — add `Day`, `Work start`, `Work end`, `Break start`,
  `Break end`, `No exceptions`; replace `Special` + `hours` with the single key
  `Special hours`.
- **"Behavior" → weekly grid** — say the grid has a labelled header row and that
  an inactive day's four selects are `disabled`.

Do not restate the day-index rule anywhere else in the repo. Two copies drift
and someone follows the stale one.

### 7.2 Update the existing tests to the new API

`scheduleeditor_test.go` and `scheduleeditor_ui_wasm_test.go` construct
`WeeklyRow{DayOfWeek: …}` and `OnWeeklyChange: func(r WeeklyRow)`. Migrate every
occurrence. Find them all:

```
grep -rn "DayOfWeek\|OnWeeklyChange" scheduleeditor/
```

### 7.3 The tests that would have caught the bug

Add to `scheduleeditor_test.go`:

```go
// The seven rows render the seven day names in order. The bug this replaces:
// the host left DayOfWeek at its zero value on unconfigured days and four rows
// rendered "Sunday".
func TestWeekRendersSevenDistinctDaysInOrder(t *testing.T) {
	e := &ScheduleEditor{Week: make([]WeeklyRow, 7)}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	for i := 0; i < 7; i++ {
		want := date.WeekdayName(i)
		if !strings.Contains(html, want) {
			t.Errorf("row %d: missing day name %q:\n%s", i, want, html)
		}
	}
}

// OnWeeklyChange reports the row's POSITION, whatever the row holds. This is
// the assertion that makes "enable Tuesday, save Sunday" unrepresentable.
func TestWeeklyChangeReportsThePosition(t *testing.T) {
	var gotDay int
	var gotRow WeeklyRow
	e := &ScheduleEditor{
		Week:           make([]WeeklyRow, 7),
		OnWeeklyChange: func(d int, r WeeklyRow) { gotDay, gotRow = d, r },
	}
	e.Init(&emptyCtx{})

	e.weeklyChange(2, func(r *WeeklyRow) { r.Active = true })

	if gotDay != 2 {
		t.Errorf("dayOfWeek = %d, want 2 (Tuesday)", gotDay)
	}
	if !gotRow.Active {
		t.Error("the mutation did not reach the reported row")
	}
}

// An inactive day's four time selects are disabled: they are not a schedule,
// they are the absence of one.
func TestInactiveDayDisablesItsTimeSelects(t *testing.T) {
	week := make([]WeeklyRow, 7)
	week[1] = WeeklyRow{Active: true, WorkStart: 480, WorkFinish: 840}
	e := &ScheduleEditor{Week: week}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	// 7 rows x 4 selects = 28; row 1 is active, so 24 are disabled.
	// Count the whole attribute, not the bare word: dom serializes as
	// key='value' (single quotes), so "disabled" alone appears TWICE per
	// select and would score 48.
	if got := strings.Count(html, "disabled='disabled'"); got != 24 {
		t.Errorf("disabled selects = %d, want 24:\n%s", got, html)
	}
}

// The header labels every column. Four unlabelled dropdowns was the report.
func TestWeekHeadLabelsEveryColumn(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, string(clsWeekHead)) {
		t.Errorf("the weekly grid has no header row:\n%s", html)
	}
	for _, k := range weekHeadKeys {
		if !strings.Contains(html, k) {
			t.Errorf("header missing column %q:\n%s", k, html)
		}
	}
}
```

Two verified facts these tests rely on:

- **`widget.Class` is a `string` type** (`webtyp.com/widget/widget.go:15`), so
  `string(clsWeekHead)` is a plain conversion and `clsWeekHead.String()` is the
  equivalent. Either is fine; be consistent with the file, which already
  asserts on the literal `"scheduleeditor__week"`.
- **`scheduleeditor_test.go` already imports `strings`, `testing` and
  `webtyp.com/date`**, and carries `//go:build !wasm`. Add no import for these
  tests. Production code in this package must still never import `strings`.

---

## 8. Constraints — read before writing code

- **No standard library in WASM-compiled code.** `webtyp.com/fmt`, never
  `strconv`, `strings`, `errors`. `_test.go` files are the only exception, and
  this package's tests already import `strings` and `testing`.
- **`dom.Element` is embedded by VALUE** — `Element`, never `*Element`. It is
  already correct in `ScheduleEditor`; do not change it.
- **SSR split by extension.** `css.go` carries `//go:build !wasm` and holds
  `RenderCSS`. Never move CSS into `scheduleeditor.go`, and never create
  `ssr.go` or `front.go` — both conventions are eliminated in this ecosystem.
- **Never compose an id string inside `Render()`.** Use `Key(...)` +
  `el.Ref()`, and `data-*` as the breadcrumb for tests and CSS. The
  `data-day` attribute in §3.4 is the sanctioned form.
- **No `map`** in code that reaches the WASM binary — it inflates the TinyGo
  output. The `[]string` and `[]fmt.KeyValue` slices used above are deliberate.
- **No `TODO`, no commented-out block, no deprecated field kept "for one
  version".** `WeeklyRow.DayOfWeek` is deleted, not deprecated. Before closing:
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' scheduleeditor/` — every
  hit must predate this change.
- Run `gotest`, never `go test`.

---

## 9. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./scheduleeditor/` | green, including the four new tests and the `StateAttrs` test |
| 2 | `gotest ./...` | green — `TestKindAllowsEveryState` and `TestNoRemovedSymbols` in `conformance_test.go` included |
| 3 | `grep -rn "DayOfWeek" scheduleeditor/` | **empty** |
| 4 | `grep -rn "style.Grow()" scheduleeditor/css.go` | **empty** |
| 5 | `grep -n "PartExcHours" scheduleeditor/css.go` | one `Part(PartExcHours, …)` rule with `RevealedBy` |
| 6 | `grep -n "RevealedBy" scheduleeditor/css.go` | exactly two hits: `PartExcForm`, `PartExcHours` |
| 7 | `grep -rn 'Translate("Special"' scheduleeditor/` | **empty** |
| 8 | `grep -rn 'Translate("Special hours")' scheduleeditor/` | one hit |
| 9 | `GOOS=js GOARCH=wasm go build ./...` | compiles |
| 10 | `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' scheduleeditor/` | no hit introduced here |
| 11 | `README.md` | `DayOfWeek` absent; new keys listed; two-parameter callback shown |

---

## 10. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | The day is the index | `scheduleeditor.go` | `DayOfWeek` deleted; `OnWeeklyChange(int, WeeklyRow)`; label and `Log` read `i`; `weekDays` guard |
| 2 | The reveal reveals | `css.go`, `scheduleeditor_test.go` | `RevealedBy(widget.Open)` on both parts; `StateAttrs` test green |
| 3 | A legible grid | `scheduleeditor.go`, `css.go` | header row; `FixedGrid(5)`; `Grow()` gone; inactive days disabled; tap targets |
| 4 | One translation key | `scheduleeditor.go` | `"Special hours"` |
| 5 | README + tests | `README.md`, both `_test.go` | every check in §9 passes |

## 11. Residual — what shipped wrong, and why

**Status: executed 2026-09-08, published as `components v0.6.22` (`7fd7300`).**
Verified live: seven distinct day names in order, `weeklyChange` reports the
position (activating Martes no longer writes Domingo), the exception form is
`display: none` with no day picked, the hour row hides for `HOLIDAY`, the
columns are labelled, and the type label reads "Horario especial".

Two defects remain, both traceable to this document rather than to its executor.

### 11.1 `style.Center()` in `PartDay` — an error in §5.4 as first written

The rule shipped as:

```go
		Part(PartDay,
			style.Row(style.Space2),
			style.Center(),          // ← wrong; §5.4 has since been corrected
		).
```

`Row` and `Center` assign the same `rule.flowType`, so `Center()` won and the
day cell is not a flex row at all. Measured live: `.scheduleeditor__day`
computes `display: block`, `max-width: 584px`, `align-items: normal`; the day
name renders at `y=296` while its checkbox sits at `y=262` — a **34px vertical
misalignment** inside a 59px row.

**Fix: delete the `style.Center()` line** from `css.go`. Nothing replaces it —
`Row(style.Space2)` already emits `align-items: center`.

**Acceptance:** `getComputedStyle('.scheduleeditor__day').display === 'flex'`,
and the day name and its checkbox share a common vertical centre.

### 11.2 "No break" renders as 06:00

`hourOptions` floors at 360 (06:00), so the `0` that `appointment_booking` uses
for "no break" has no matching `<option>` and the select falls back to its
first. Measured live: a Monday seeded with **no break** shows
`Colación desde 06:00 / Colación hasta 06:00`.

This contradicts this component's own README — "the break is shown and used only
when both break fields are non-zero" — and it predates this plan. Labelling the
columns did not cause it, but it made it legible: an unlabelled 06:00 was
meaningless, a "Colación desde 06:00" is a false statement.

It is not currently corrupting (`weeklyChange` copies the stored row and mutates
only the touched field, so an untouched `0/0` survives), but touching either
break select persists a real 06:00 break.

**Fix — pick one, and say which in the follow-up plan:** render the two break
selects only when the row has a break, or give them an explicit "no break"
option carrying the value `0`. The second keeps the grid rectangular and is
probably right, but it is a UI-copy decision, not one to make silently.

## 12. Note for the reviewer — what is NOT in this plan

`.scheduleeditor` renders **1,950 `<option>` elements** (7 rows × 4 selects × 65
options, plus the exception form's 2 × 65) — 55% of the demo page's 3,524 nodes.
The native `<select>` is kept by an explicit owner decision; replacing it with a
lighter range control is a separate API change and is deliberately **out of
scope**. Recorded so it is not rediscovered as a new finding.
