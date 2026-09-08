# webtyp/components — Architecture

WHAT & WHY, not HOW. Implementation guidance lives in
[`components/AGENTS.md`](../../AGENTS.md) and the component skill
[`SKILL.md`](./SKILL.md).

## The shape of a component

A component is a leaf building block with zero layout knowledge: it implements
`Render() *dom.Element` (plus an optional `Init(ctx)`) and declares its visual
contract through `RenderCSS()` in a `//go:build !wasm` file. Layout skeletons
live in `webtyp/layout`; this repo never imports it. See
[`README.md#where-components-fit`](../README.md#where-components-fit).

```
flowchart TD
  A[consumer composition root] --> B[webtyp/layout shell e.g. crudview]
  B --> C[webtyp/components leaf]
  C --> D[typed signals + Bind*]
  C --> E[RenderCSS&#40;&#41; sheet]
```

## Lego pieces — assemblies own the concern, never the copy

Some concepts are shared between leaf components and cannot be re-declared
per widget or the copies drift. `listgap` (list scroll gutter) and
`listselect` (multi-selection mode) are such lego packages: a widget imports
and assembles them; the piece owns the DOM and the skin.

- `listselect` owns **the selection chrome**, not just the mode state: the
  per-row check box and the in-flow select-all header strip (box +
  `k / N` count). `RowOf`/`Header` build the markup, `ApplyRow`/`ApplyHeader`
  paint it; `targetlist`, `targetdate` and `targethour` `Child()` the pieces
  into their rows and roots and never write their own version. The header is
  in flow (a normal row above the `<ul>`), hidden in normal mode — nothing is
  absolutely positioned over the first row, so a checked list never overlaps
  its content.

## Navigation that must not own `location.hash`

`calendarslider` moves between neighbouring months with scroll-snap. The
controls are `<button>`s whose click handler (`slideToMonth`) jumps the strip
with `ScrollIntoView` — deliberately **not** `<a href="#cs-m-…">` anchors.
An anchor mutates `location.hash`, and a hash-routed shell (`platformd`)
reads that as a route change and blanks the view; a button keeps the slide
inside the widget.

## `scheduleeditor` — a pure editor that composes another component

`scheduleeditor` is a leaf component that edits a professional's availability
through three mechanisms shown at once: a **weekly pattern** (`Pattern
[]PatternRow` — time ranges tagged with weekday chips), **marked working days**
(`Marked []MarkedDay` — concrete dates), and **per-date exceptions**
(`Exceptions []Exception` — closed / special hours / blocked). It is **pure**:
it knows nothing of `router`, `orm` or any domain module — the host feeds it
that state plus `Bounds`/`Holidays`/`Closures` and translates its callbacks to
persistence ops. That split keeps it usable by any host.

It composes `calendarslider` **twice** — once collapsed for the bulk day marker
(multi-select via `SelectedMany`/`OnToggle`) and once for the exceptions panel
(single-select via `Selected`/`OnSelect`) — mapping the component's
`Holidays`/`Closures []string` → `calendarslider.Holiday` and its `Marked` /
`Exceptions` → the `Occupation` list that makes a day selectable. The assembled
widget consumes a sibling leaf, it never re-declares its calendar; multi-select
was added to `calendarslider` itself rather than forked here. See
[`scheduleeditor/README.md`](../scheduleeditor/README.md).