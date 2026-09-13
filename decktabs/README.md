# DeckTabs

A strip of tabs over a deck of panels: one region of a screen holding several
sections, with one on view at a time.

[← Back to Catalog](../docs/CATALOG.md)

## What it is

`DeckTabs` gives a screen more than one section without giving the application more
than one route. It is the widget-scale version of the mechanism
`webtyp/layout`'s `platformd` already runs at page scale: the panels are stacked
as layers by `style.SlideDeck`, and whichever one carries the `widget.Current`
state is the one on screen.

It composes only existing vocabulary — `widget.Tabs` (the Kind),
`widget.Current` (the state) and `style.SlideDeck` (the reveal). There is no
new reveal mechanism here.

## Usage

```go
import (
    "webtyp.com/components/decktabs"
    . "webtyp.com/dom"
)

active := NewString("") // yours to keep: read it, set it, persist it

t := &decktabs.DeckTabs{
    Label:  "Personal",          // becomes the tab strip's aria-label
    Active: active,              // optional; nil gets one, defaulting to Items[0]
    Items: []decktabs.Item{
        {ID: "data",     Label: "Datos",     Icon: iconStaff, Panel: dataView},
        {ID: "hours",    Label: "Horario",                    Panel: hoursView},
        {ID: "services", Label: "Servicios",                  Panel: servicesView},
    },
    OnChange: func(id string) { /* reload, record, navigate */ },
}
```

## Fields

| Field | Meaning |
|---|---|
| `Items` | one `Item` per section. `Item.ID` must be unique within the set and stable across renders — it is what `Active` carries and what the ARIA attributes link by |
| `Active` | `*dom.SignalString` holding the ID on view. Optional; `nil` gets one created on `Init`, defaulting to `Items[0].ID` |
| `Label` | optional; becomes the strip's `aria-label`, which a screen reader announces as the name of the tab set |
| `OnChange` | fires **after** `Active` changes. Re-clicking the tab already on view does not fire it |

`Item.Panel` is any `dom.Component`. `Item.Icon` is optional; the empty icon
renders none. Labels are consumer-supplied text — this component translates
nothing.

## Three things to get right

1. **Every panel stays mounted.** That is `SlideDeck`'s contract: the state
   decides what is on screen, nothing is unmounted. A panel may hold a form
   mid-edit and survive a trip through another tab. Do not expect a panel to be
   constructed lazily when its tab is first opened.
2. **The strip wraps; it never scrolls.** A horizontal scroller here would chain
   with the horizontal scroll a panel's own content may have, and a swipe inside
   the content would end up changing tab on its own. If a set has too many tabs
   to fit, it has too many tabs.
3. **`Active` is yours.** It is injected rather than owned so the active tab can
   be driven from outside — a URL fragment, a saved preference, a sibling
   control that jumps to a section. Read it and set it freely; the component
   re-reads it on every render.

## Accessibility

The strip is `role="tablist"`, each tab `role="tab"` with
`aria-controls` pointing at its panel, and each panel `role="tabpanel"` with
`aria-labelledby` pointing back. Tabs render as `<button type="button">` — a
bare `<button>` defaults to `type="submit"`, which would submit the surrounding
form on every tab change.

Those ARIA references use ids **minted by `dom`** (`Element.GetID()`), never ids
composed from `Item.ID`. That is why the same `Item.ID` may appear in several
tab sets on one page: `dom` resolves handlers and signal patches by id, and a
duplicate id silently kills one of the two nodes. `Item.ID` is still rendered as
`data-id` if you need it from a query or a test.
