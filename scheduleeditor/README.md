# ScheduleEditor

A pure component for editing a professional's **weekly schedule template**
(7 rows: Sunday..Saturday) plus **per-date exceptions** (closed / special hours
/ blocked). Pure means it knows nothing about `router`, `orm` or any domain
module — the host supplies the data (initial state) and translates the
callbacks to persistence ops.

## Import

`"webtyp.com/components/scheduleeditor"`

## Data shape

Times are **minutes from midnight** (`int`, 0..1439) — the same shape the
`appointment_booking` module persists in `work_calendar_weekly`. Time `0/0`
for the break means "no break".

```go
type WeeklyRow struct {
    DayOfWeek               int  // 0=Sunday .. 6=Saturday
    Active                  bool
    WorkStart, WorkFinish   int  // minutes from midnight
    BreakStart, BreakFinish int  // 0/0 = no break
}

type Exception struct {
    ID   string // "" when adding (host assigns on persist)
    Date string // "YYYY-MM-DD"
    Type string // ExcHoliday | ExcSpecialHours | ExcBlocked
    StartMin, EndMin int // used by SpecialHours / Blocked
    Notes string
}
```

Exception types are the exported constants `ExcHoliday`, `ExcSpecialHours`,
`ExcBlocked` — never inline strings.

## Usage

```go
editor := &scheduleeditor.ScheduleEditor{
    Week:      rowsFromHost,          // 7 rows Sun..Sat
    Exceptions: exceptionsFromHost,   // sorted by date asc
    Holidays:  nationalHolidays,      // "YYYY-MM-DD", read-only
    OnWeeklyChange: func(r scheduleeditor.WeeklyRow) {
        _ = client.SaveWeeklyRow(toAppointment(r)) // host persists, then reloads
    },
    OnExceptionAdd: func(ex scheduleeditor.Exception) {
        _ = client.AddException(toAppointment(ex))
    },
    OnExceptionRemove: func(id string) {
        _ = client.RemoveException(id)
    },
}
```

`Week`/`Exceptions` are **initial state**: the host persists on each callback
and re-mounts with fresh data (same model as `targethour`/`crudview` — the
component is not the source of truth).

## Behavior

- **Weekly grid** — one row per day; a checkbox enables the day; four selected
  hour selectors (`work-start`, `work-finish`, `break-start`, `break-finish`)
  offer 06:00–22:00 in 15-minute steps (values are minutes). The break is
  shown and used only when both break fields are non-zero. An invalid row
  (work window, or break outside work) is marked `data-invalid` and logged as a
  dev warning — it never blocks the edit.
- **Exceptions panel** — a `calendarslider` whose selectable days are the dates
  with exceptions (occupation 100 for holiday/blocked, 50 for special hours).
  Picking a day reveals an inline add form: type (Closed / Special hours /
  Blocked), optional hours for the two time-based types, optional notes.
  "Add" fires `OnExceptionAdd` with `ID == ""`. The vigente list sorts by date,
  shows a "Remove" button, and renders national `Holidays` read-only.

## Translation keys

The component renders its chrome through
[`webtyp.com/fmt/lang`](https://pkg.go.dev/webtyp.com/fmt/lang). It registers
**no dictionary** — the consumer's does. Keys introduced:

- Day names via `date.WeekdayName` (the 7 canonical English names
  `Sunday`..`Saturday`).
- Type labels: `Closed`, `Special hours`, `Blocked`.
- Chrome: `Add`, `Remove`, `Type`, `Date`, `Notes`.

## Tests

`gotest ./scheduleeditor/` — backend (`RenderCSS`, row/render structure,
callbacks) plus a real-DOM `wasm` interaction test (day pick reveals the form).

Publish with `gopush 'message'`.