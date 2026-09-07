# Components Catalog

This catalog documents the available reusable components in `webtyp/components`.

[← Back to Main README](../README.md)

## Theme

All components consume CSS custom properties from `webtyp/dom`'s `theme.css`.
Inject `dom.ThemeCSS` into your page `<head>` once via the site builder.
Components do not define colors — they inherit from the theme.

## Where components fit

`components` are raw reusable pieces with no layout knowledge. A layout
skeleton in `webtyp/layout` (`platformd`, `rightpanel`, `crudview`) exposes
named slots; the consumer's composition root (e.g. `config/layouts`)
preconfigures a layout once and injects these components into its slots.
Modules never assemble components by hand, and this repo never imports
`webtyp/layout`. See the [root README](../README.md#where-components-fit)
for the full diagram.

## Overview

All components follow the [Component Creation Guide](./SKILL.md). Every entry
below is **slot-ready**: it satisfies `dom.Component`, has a passing
`*_contract_test.go`, and is theme-driven (no hardcoded colors) per the
Stage 1/2 contract in [SKILL.md](./SKILL.md).

---

## [ActionButton](../actionbutton/README.md) — ✅ Slot-ready
Versatile button component with variant support (primary, secondary, danger).
[Detailed Documentation →](../actionbutton/README.md)

---

## [CountBadge](../countbadge/README.md) — ✅ Slot-ready
Notification bubble riding the host button's top-end corner, out of the flow
so the host keeps its box at any count. Amber (`AccentInverse`) pill, shown
only above zero.
[Detailed Documentation →](../countbadge/README.md)

---

## [ContentCard](../contentcard/README.md) — ✅ Slot-ready
Container with header, body, and footer sections.
[Detailed Documentation →](../contentcard/README.md)

---

## [CalendarSlider](../calendarslider/README.md) — ✅ Slot-ready
Declarative month calendar (zero JS): a window of `NumMonths` months centered
on the current month, sliding between them via scroll-snap. The ‹ › controls
are buttons with a small `slideToMonth` handler (never `<a href="#cs-m-...">`
anchors — an anchor mutates `location.hash`, which a hash-routed shell reads
as a route change). Holidays (red), occupation percentage (selectable days
with `N%` bar), today marker, and single-day selection through a
`*dom.SignalString` (`YYYY-MM-DD`) with `BindState(widget.Selected, ...)` —
the DOM patches in place, never re-renders. Day cells fill their grid track
(squares via aspect-ratio, ~24px floor, no fixed boxes), so the grid always
uses the full card width. The month nav is one static
centered row (‹ label ›, 50px boxes from the closed scales), identical on
desktop and touch — no hover reveal anywhere. Folding works on every
viewport: the strip shows if and only if expanded (desktop defaults to
expanded), and the calendar chip is a full-width bottom bar under the
strip — always visible as the fold toggle, never floating, never
overlapping. The swap fades via `Animate(MotionSlow)` paired with
`RevealedBy` (widget v0.6.22), instant under `prefers-reduced-motion`.
Satisfies `widget.Filterable` for `crudview` filter integration.
[Detailed Documentation →](../calendarslider/README.md)

---

## [ModalDialog](../modaldialog/README.md) — ✅ Slot-ready
Centered modal overlay with backdrop and close button. Named for its specific
style (centered/backdrop) so other dialog styles (e.g. a drawer or a confirm
prompt) can be added later as their own two-word component.
[Detailed Documentation →](../modaldialog/README.md)

---

## [DataTable](../datatable/README.md) — ✅ Slot-ready
Data table for structured information with headers and rows.
[Detailed Documentation →](../datatable/README.md)

---

## [SelectSearch](../selectsearch/README.md) — ✅ Slot-ready
Signal-driven searchable dropdown with static options, live filtering, and optional DB search callback. Satisfies `widget.Filterable` — drop-in for any host slot that accepts a filter control (same seam `searchbar.SearchBar` fills). Uses `BindChildren` for efficient list updates and `Show` for the dropdown.
[Detailed Documentation →](../selectsearch/README.md)

---

## [SearchBar](../searchbar/README.md) — ✅ Slot-ready
One-control filter bar: a magnifier cap and a text field sized by
`--control-height`. Reports each keystroke through `OnFilterChange(term)` and
knows nothing about what it filters, so a host can swap it for a calendar or a
select in the same slot.
[Detailed Documentation →](../searchbar/README.md)

---

## [ScheduleEditor](../scheduleeditor/README.md) — ✅ Slot-ready
Pure editor for a professional's weekly schedule template (7 rows: Sunday..
Saturday with enable toggle + 4 hour selects) plus per-date exceptions (closed
/ special hours / blocked) over a `CalendarSlider`. Times are minutes from
midnight — the shape `appointment_booking` persists. The host feeds the data
and translates the callbacks to ops; the component knows nothing of
`router`/`orm`.
[Detailed Documentation →](../scheduleeditor/README.md)

---

## [TargetHour](../targethour/README.md) — ✅ Slot-ready
Selectable list component for a day's booked slots: each row leads with a prominent hour (`HH:MM`) and carries an optional per-row status tint (`pending` / `confirmed` / `attended`). Same multi-selection mechanics, selection header, and `crudview.ListView` compatibility as `targetlist` and `targetdate`. Pairs with `CalendarSlider` (now `widget.Filterable`) as the crud filter.
[Detailed Documentation →](../targethour/README.md)

---

## ListSelect — ✅ Slot-ready (lego piece)
Not a standalone component: owns the multi-selection chrome the `target*`
lists assemble. `Mode` is the selection state; `RowOf` builds a row's check
box plus its narrow Edit/Danger signals; `Header` builds the in-flow
select-all strip (box + `k / N` count) — hidden in normal mode, revealed by
the list root's `Open` state. `ApplyRow`/`ApplyHeader` paint the pieces.
`targetlist`, `targetdate` and `targethour` assemble it, never re-declare it.

---

## [ThemeToggle](../themetoggle/README.md) — ✅ Slot-ready
Signal-driven floating button that cycles between `auto → dark → light` theme modes. Persists preference in `localStorage` via `Init`. Uses derived signals for labels and icon updates.
[Detailed Documentation →](../themetoggle/README.md)

---

## Forms

Forms are NOT part of `webtyp/components`.
Use `webtyp.com/form` directly — it is the only form library in the
webtyp ecosystem and provides field layout, validation, labels and submission.
