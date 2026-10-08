# ThemeToggle

Signal-driven component to toggle the application theme (dark / light).

## Features

- 2 states: `dark` and `light`. With nothing saved it imposes neither: `<html>` is left
  alone and the icon shows the scheme in effect — the app's declared default
  (`css.Theme(css.DefaultLight())`) or, without one, the OS preference. Only a click
  writes `data-theme`.
- Automatic persistence in `localStorage` via `Init`.
- Fixed floating button.
- Signal-driven: updates icon and labels surgically without re-rendering the whole button.

## Usage

```go
import (
	"webtyp.com/components/themetoggle"
	"webtyp.com/dom"
)

func main() {
	ts := &themetoggle.ThemeToggle{}
	dom.Append("body", ts)
	select {}
}
```
