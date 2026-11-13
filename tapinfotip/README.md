# TapInfotip

A small help button (ⓘ) that shows its text only when tapped. A second tap hides it.
Nothing is drawn for the text until someone asks for it, so the help costs no space
while it is not wanted.

## Features

- The button is the shared information glyph (`webtyp/icons/circleinfo`), drawn in
  `currentColor`, inside a 44px touch target.
- `aria-expanded` follows the panel; the button's `aria-controls` names the panel.
- Opening the panel brings it into view (`block: nearest`), so a help text under the
  last field is not cut off by the screen edge.
- The text is an English key, translated when drawn (`lang.Text`).

## Usage

A form places the two halves itself: the button in the label row, the panel under the
control.

```go
import "webtyp.com/components/tapinfotip"

tip := &tapinfotip.TapInfotip{Help: "Document number, as written on the document."}
labelRow.Child(tip.Button())
field.Child(tip.Panel())
```

A free-standing use renders both in one wrapper:

```go
tip := &tapinfotip.TapInfotip{Help: "Uncheck to retire the patient; the record is never deleted."}
dom.Append("body", tip)
```

## Notes

- The name follows the two-word rule of `components`: `TapInfotip` = *how it is opened* (tap)
  + *the generic class* (infotip).
- Visual rules live in `css.go`; the glyph's geometry in `svg.go` (`!wasm`).
