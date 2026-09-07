# TargetHour

`targetlist`'s sibling for a day's booked slots: each row leads with a
prominent hour (`HH:MM`) and may carry a status tint (pending / confirmed /
attended). Same multi-selection mechanics as `targetlist` / `targetdate` — it
assembles `components/listselect`, it does not re-declare it — so a `crudview`
can swap one for the other as `Config.List`.

## Import
`"webtyp.com/components/targethour"`

## Usage

```go
import "webtyp.com/components/targethour"

th := &targethour.TargetHour{
	Selected: selected,        // *dom.SignalString shared with the crud form (optional)
	OnSelect: func(it view.Item) { /* row tapped in normal mode */ },
	StatusOf: func(it view.Item) targethour.Status {
		// host maps its own model / view.Item to the typed status enum
		switch it.Description {
		case "Confirmada":
			return targethour.StatusConfirmed
		case "Atendida":
			return targethour.StatusAttended
		}
		return targethour.StatusPending
	},
}
th.SetItems(items) // []view.Item — each with LeadMain set to the hour "HH:MM"
```

### Free slots (reservable hours)

`FreeSlots []string` (hours `"HH:MM"` with no reservation) renders light, distinct
rows after the booked ones — visually "available", with a trailing `+`. Tapping
one calls `OnPickFree(hhmm)`. They are actions, not records: they never
participate in selection/delete mode. `FreeSlots` nil renders nothing new:

```go
th := &targethour.TargetHour{
	Selected:   selected,
	OnSelect:   onSelect,
	StatusOf:   statusOf,
	FreeSlots:  []string{"09:00", "10:30"}, // recomputed from list_availability
	OnPickFree: func(hhmm string) { /* start booking at hhmm */ },
}
```

As a `crudview` list:

```go
List: func(selected *dom.SignalString, onSelect func(view.Item)) crudview.ListView {
	return &targethour.TargetHour{Selected: selected, OnSelect: onSelect, StatusOf: statusOf}
},
```

## Properties

- `Selected` (`*dom.SignalString`): id of the highlighted row. Optional —
  created if nil so a host can share it (e.g. a CRUD view binding the form to
  the same signal).
- `OnSelect` (`func(view.Item)`): the row body was tapped in normal mode.
  Ignored while selection mode is on (a tap toggles the check there).
- `StatusOf` (`func(view.Item) Status`): maps a row to its booking state for
  the row tint. Optional — nil means every row is `StatusPending` (no tint).
  The host owns the mapping so the library holds no localized status strings.

## The row

`view.Item` fields read: `ID`, `LeadMain` (the hour, rendered prominent),
`Label` (the name), `Description` (an optional trailing chip, truncated to 16
chars). `LeadTop` / `LeadBottom` are ignored — use `targetdate` if you need
the stacked date badge.

## Status tint

| `Status` | Row state written | Meaning |
|---|---|---|
| `StatusPending` (zero value) | none | no tint — a plain row |
| `StatusConfirmed` | `data-locked="true"` | confirmed by reception (`AccentWash`) |
| `StatusAttended` | `data-busy="true"` | patient already attended (`Subtle`, muted) |

The tint is a low-key wash so it never fights the blue selection highlight.
The state is captured per row when `SetItems` builds it — the host rebuilds
rows on every reload, so there is no signal.

## Selection mode

Identical to `targetlist` / `targetdate`: `SetSelectMode(true)` opens it (the
root carries `data-open`), a tap toggles a row's check, `SetDanger(true)`
paints checked rows red (delete) instead of blue (edit), `CheckedIDs()`
returns the marked ids in render order, `OnCheckedChange(fn)` fires the count.
The check glyph (trash / pencil) comes from `webtyp.com/icons`, the
same one the crud view's footer buttons draw.

### Selection header (select all)

In selection mode, `listselect`'s header strip appears above the list: a
select-all box and an `n / total` count. It operates as a tri-state toggle:
tapping it when no rows or some rows are checked selects all rows; tapping it
when all rows are checked clears the selection. The box carries the mode glyph
(trash in danger mode, pencil otherwise) on a solid fill — nothing is
absolutely positioned over the rows, so the first row is never overlapped.

## `crudview` integration

Pair it with `calendarslider.CalendarSlider` as the `crudview.Config.Filter`:
`CalendarSlider` satisfies `widget.Filterable`, so picking a day emits that
date (`"YYYY-MM-DD"`) as the filter term and the presenter's `Filter(term)`
returns that day's slots.
