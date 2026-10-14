# ModalDialog

Modal dialog overlay with backdrop and close button. The title uses `style.Heading`.

## Import
`"webtyp.com/components/modaldialog"`

## Usage

```go
import (
    "webtyp.com/components/modaldialog"
    . "webtyp.com/html"
)

m := &modaldialog.ModalDialog{
    Title: "Confirm Action",
    Content: P("Are you sure?"),
}

// To open
m.Open()
```

## Click model

The root of the dialog IS the wash: it covers the viewport (`Backdrop(Viewport)` +
`Veil`) and is the click target — a click on it closes the dialog. The panel is
the root's in-flow child, so it always paints above the wash and never needs a
stacking level; it stops the click from reaching the root, so interacting with
the dialog never closes it. There is no separate click-catcher element (the
`PartBackdrop` part was removed — it was the one sibling the DSL could not
place below an in-flow peer).

## Properties
- `Title` (string): Modal title.
- `Content` (dom.Component): Modal body content.

## Methods
- `Open()`: Shows the modal.

---
[Back to Catalog](../docs/CATALOG.md)
