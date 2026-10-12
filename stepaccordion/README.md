# StepAccordion

A vertical wizard. Steps open one at a time, in order: a finished step collapses to a
one-line summary with an **Edit** button, and every step below the open one stays locked
until it is reached. Reopening a step locks everything after it again and calls `OnReopen`,
so the host can clear the choices that depended on it.

Each step's phase is written on its `<section>` as a state — open (`widget.Open`),
finished (`widget.Locked`) or locked (`widget.Disabled`) — and the stylesheet reveals the
body, summary and Edit button from it.

## Usage

```go
import (
	"webtyp.com/components/stepaccordion"
	. "webtyp.com/html"
)

acc := &stepaccordion.StepAccordion{
	Steps: []stepaccordion.Step{
		{Key: "service", Title: "Service", Body: servicePicker},
		{Key: "slot", Title: "Date and time", Body: slotPicker},
		{Key: "summary", Title: "Summary", Body: P().Text("…")},
	},
	OnReopen: func(key string) { clearChoicesAfter(key) },
}

// inside the service picker, once the user picks:
acc.Advance("Consulta general")
```

- `Advance(summary)` finishes the open step and opens the next one (none after the last).
- `Reopen(key)` reopens a finished step; unknown or not-finished keys are ignored.
- `Open()` returns the key of the open step, or `""`.

Titles are `lang.Text` (English keys, translated in `Render`); the "Edit" label too.
