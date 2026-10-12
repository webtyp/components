# ChipGrid

A grid of chips to pick one value — for example the free hours of a day. It is a WAI-ARIA
listbox: a single tab stop (the selected chip, else the first enabled one), arrows and
Home/End move between enabled chips, Enter or Space pick. Disabled chips are shown but
skipped and cannot be picked.

## Usage

```go
import "webtyp.com/components/chipgrid"

grid := &chipgrid.ChipGrid{
	Empty:  "No times available", // lang.Text; default "No options available"
	OnPick: func(hour string) { book(hour) },
}
grid.SetChips([]chipgrid.Chip{
	{Value: "09:00", Label: "09:00"},
	{Value: "09:30", Label: "09:30", Disabled: true},
	{Value: "10:00", Label: "10:00"},
})
```

- `Selected` is a `*dom.SignalString` the host may share; it holds the picked value.
- `SetChips` replaces the chips and clears `Selected` when its value is no longer an
  enabled chip.
- `Label` is data and is shown as is; only `Empty` is translated.
