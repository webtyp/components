# CalendarSlider

Declarative, JavaScript-free month calendar: one month visible at a time, sliding to its neighbors, with holidays, occupation percentage, today marker, and single-day selection — all signal-driven.

## Features

- One month on screen at a time, starting at `Start` (default: today's month) and sliding forward through up to `NumMonths` (max 12) consecutive months. ‹ › are `<button>`s; a small click handler (`slideToMonth`) jumps the scroll-snap strip to the neighbouring month card with `ScrollIntoView`, and the browser's scroll-snap animates it. They are deliberately **not** `<a href="#cs-m-...">`: an anchor mutates `location.hash`, which a hash-routed shell (`platformd`) reads as a route change, blanking the view. No rebuild, no JS.
- Week starts Monday; Sunday and holidays render red (Danger). Each month card opens with its `‹ month year ›` header, then the weekday row, then the grid — the label names the grid underneath it, so it is read first.
- Occupation days (`Occupation []OccupationDay`) are selectable and show a `N%` bar (`data-use`) whose **colour is the level**, not just its length: Success under 50%, DangerWash to 84%, Danger at 85% and over. A bar with no track behind it cannot be read as a fraction by length alone. Values clamp to 0–100, and the percentage goes into the day's `aria-label`, not only `title`.
- Today's date is marked with a Today style and `title` from the `Today` key.
- Selected day is a `*dom.SignalString` (`YYYY-MM-DD`) written by the host or by clicking a bookable day; the DOM patches in place via `BindState(widget.Selected, ...)` — no full re-render.
- `OnSelect` callback fired on click, in addition to the signal.
- A selectable day is a real `<button>` inside its `role=gridcell`, so Enter/Space select it — no JS. Roving tabindex: exactly one day per month card carries `tabindex="0"` (selected, else today, else first selectable day), the rest `-1`, so Tab enters and leaves each card in one stop. Arrow keys move inside the card (`ArrowLeft/Right` ±1 day, `ArrowUp/Down` ±7, `Home/End` to the week's ends) via `dom.OnKeyDown`. Blocked days render a plain `<div>` and are not focusable, because there is nothing to activate. The button fills the whole cell, so the hit area stays the full ~46px square.
- Accessible: `role=grid`/`row`/`gridcell`/`columnheader`, `aria-selected`, `aria-hidden` filler cells, labeled navigation buttons.
- Light/dark out of the box: every color comes from theme tokens (`light-dark()`-aware), so the calendar follows whatever `data-theme` the app sets — pair it with `webtyp/components/themetoggle` for a user-facing switch.
- The collapsed field is the component's **first** row: a `--control-height` (50px) bar showing the chosen date, or a `Select date` placeholder. It is the trigger, and the month panel hangs off it — the shape every date picker has.
- **One size everywhere, no breakpoints.** The root caps at `Center(Compact)` = `min(100%, 24rem)`, so the day cell lands at ~46px — above the 44px touch floor — on a phone and on a 4K screen alike. A day cell is a *control*, not content: it holds a number and a 4px meter, so nothing inside it benefits from more room. Agenda views (Google Calendar, FullCalendar) grow their cells because each holds events; pickers (native `input[type=date]`, Material, Ant) keep one fixed box. This is a picker. Do not "free" the width to fill a wide column — that is what made the cells 164px squares and the card taller than the viewport.

## Usage

```go
import "webtyp.com/components/calendarslider"

cal := &calendarslider.CalendarSlider{
    Start:     "2026-08", // first month of the strip (defaults to today's month)
    NumMonths: 3,         // how many months forward from Start are slidable (max 12)
    Holidays:   []calendarslider.Holiday{{Date: "2026-08-15", Name: "Asunción de la Virgen"}},
    Occupation: []calendarslider.OccupationDay{{Date: "2026-08-11", Percent: 60}},
    OnSelect: func(date string) {
        fmt.Printf("selected: %s\n", date)
    },
}
cal.Init(dom.NewCtx(...))
dom.Render("app", cal.Render())

// Host-driven selection:
cal.Selected.Set("2026-08-11")
```

## API

### CalendarSlider Struct

- `Start string`: Month key `YYYY-MM`, the first (leftmost) month of the strip; empty means the current month. There is nothing to slide to before `Start` — like the original, this is built for booking forward, not browsing past months.
- `NumMonths int`: How many consecutive months forward from `Start` are slidable (default 3, max 12).
- `Holidays []Holiday`: `{Date, Name string}`; `Date` is `YYYY-MM-DD`, shown as the day's `title` and rendered red. Slice, not a map — TinyGo.
- `Occupation []OccupationDay`: `{Date string; Percent int}`; `Percent` clamps to 0–100. A date's presence in the list makes that day selectable. Slice, not a map — TinyGo.
- `Selected *dom.SignalString`: Two-way selection signal (`YYYY-MM-DD`); write it to select programmatically.
- `OnSelect func(date string)`: Callback fired when a bookable day is clicked.

### Signals

- `Init` seeds `Selected` with `dom.NewString("")` when nil — safe to use before any interaction.
- `Expanded *dom.SignalBool` decides whether the month strip is open; `Init` seeds it to `true`. Write it to open or fold the calendar from the host — e.g. seed it `false` where the calendar is a filter and vertical space is scarce.

## Translation keys

The component renders its chrome through
[`webtyp.com/fmt/lang`](https://pkg.go.dev/webtyp.com/fmt/lang). It registers
**no dictionary** — the consumer's does. Keys introduced:

- Month names (`January` … `December`) and short weekday names (`Sun` … `Sat`).
- Long weekday names (`Sunday` … `Saturday`) for the collapsed field's label.
- Chrome: `Calendar` (the strip's accessible name), `Today`, `Previous month`,
  `Next month`, `Select date` (the collapsed placeholder), `occupied` (the
  day's `aria-label` suffix).

## Screenshots

The `web/` demo (`Start: "2026-08"`, the same sample holidays/occupation shown in Usage, plus `themetoggle`) in all four combinations of theme and viewport:

| Light | Dark |
|---|---|
| ![Light, desktop](docs/screenshots/light-desktop.png) | ![Dark, desktop](docs/screenshots/dark-desktop.png) |
| ![Light, mobile](docs/screenshots/light-mobile.png) | ![Dark, mobile](docs/screenshots/dark-mobile.png) |
