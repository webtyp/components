# ScheduleEditor

A pure component for editing a professional's **weekly pattern** (time ranges
tagged with weekday chips), **marked working days** (concrete dates worked), and
**per-date exceptions** (closed / special hours / blocked). Pure means it knows
nothing about `router`, `orm` or any domain module — the host supplies the data
(initial state) and translates callbacks to persistence ops.

## Import

`"webtyp.com/components/scheduleeditor"`

## Data shape

Times are **minutes from midnight** (`int`, 0..1439).

```go
type PatternRow struct {
    StartMin, EndMin int   // minutes from midnight
    Days             []int // 0=Sunday … 6=Saturday
}

type MarkedDay struct {
    Date             string // "YYYY-MM-DD"
    StartMin, EndMin int
}

type Bounds struct {
    OpenMin, CloseMin int // 0,0 = unbounded (00:00..23:45 fallback)
}

type Exception struct {
    ID               string // "" when adding (host assigns on persist)
    Date             string // "YYYY-MM-DD"
    Type             string // ExcHoliday | ExcSpecialHours | ExcBlocked
    StartMin, EndMin int    // used by SpecialHours / Blocked
    Notes            string
}
```

Exception types are the exported constants `ExcHoliday`, `ExcSpecialHours`,
`ExcBlocked` — never inline strings.

## Usage

```go
editor := &scheduleeditor.ScheduleEditor{
    Pattern:    patternFromHost,       // []PatternRow
    Marked:     markedFromHost,        // []MarkedDay, sorted ascending
    Bounds:     Bounds{OpenMin: 480, CloseMin: 1200}, // establishment opening window
    Horizon:    6,                      // months to show in day marker
    Holidays:   nationalHolidays,      // "YYYY-MM-DD", read-only
    Closures:   establishmentClosures, // "YYYY-MM-DD", read-only
    Exceptions: exceptionsFromHost,

    OnPatternChange: func(rows []scheduleeditor.PatternRow) {
        _ = client.SavePattern(rows)
    },
    OnDaysMarked: func(dates []string, startMin, endMin int) {
        _ = client.MarkDays(dates, startMin, endMin)
    },
    OnDaysUnmarked: func(dates []string) {
        _ = client.UnmarkDays(dates)
    },
    OnMarkedDayEdit: func(day scheduleeditor.MarkedDay) {
        _ = client.UpdateMarkedDay(day)
    },
    OnExceptionAdd: func(ex scheduleeditor.Exception) {
        _ = client.AddException(ex)
    },
    OnExceptionRemove: func(id string) {
        _ = client.RemoveException(id)
    },
}
```

`Pattern`, `Marked`, and `Exceptions` are **initial state**: the host persists on
each callback and re-mounts with fresh data.

## Behavior

- **Weekly pattern** — time range rows carrying start/end selectors and seven
  weekday chips (`Sun`..`Sat`). Several rows can cover the same weekday (e.g.
  morning and afternoon sessions); a break is simply the gap between rows. Hour
  options are clamped to `Bounds`.
- **Bulk day marker** — a `calendarslider` over `Horizon` months (collapsed by
  default) allowing date marking and unmarking with a common window or custom
  hours per date. National holidays and establishment closures render
  unavailable and non-selectable.
- **Exceptions panel** — a calendar slider showing existing exceptions and
  providing an inline form to add date exceptions (Closed / Special hours /
  Blocked) with optional notes and removal support.

## Translation keys

The component renders its chrome through
[`webtyp.com/fmt/lang`](https://pkg.go.dev/webtyp.com/fmt/lang). It registers
**no dictionary** — the consumer's does. Keys introduced:

- Short weekday names (`Sun`, `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`).
- Section titles & labels: `Weekly pattern`, `Marked days`, `Hours for marked days`.
- Buttons: `Add row`, `Remove row`, `Add`, `Remove`.
- Exception type labels: `Closed`, `Special hours`, `Blocked`.
- Empty list placeholder: `No exceptions`.
- Form fields: `Type`, `Date`, `Notes`.

## Tests

Run tests with `gotest ./scheduleeditor/ ./calendarslider/`.
