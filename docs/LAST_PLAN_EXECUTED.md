# PLAN — `decktabs`: a tab set over a mounted deck of panels

> Workflow context: skill **agents-workflow**. **Executed locally on
> 2026-09-13**, not dispatched: `docs/PLAN.md` in this repository holds a
> different, gate-blocked plan (Phase C of the typed-events wave) and was not
> touched. Three deviations from the drafted plan were found while building
> against the real API — each is recorded inline below under **DEVIATION**, so
> this file matches what shipped rather than what was first proposed.

## 1. Why

An application needs to put a professional's whole setup on one screen —
identity, working hours, services — as three sections of a single module
instead of three entries in the nav rail. There is no component for that.

The repository has 21 components; none is a tab set. So today the choice is
either three separate nav entries (the operator jumps between screens to
configure one person) or a locally hand-composed reveal inside the application
— which is the leaf fix this ecosystem forbids: a repeated visual decision gets
one home, not N patched consumers.

## 2. Why this is assembly, not invention

Every piece already exists and is already used at page scale. This component is
their first use at widget scale; it adds no new vocabulary.

| Piece | Where it already lives | What it does here |
|---|---|---|
| `widget.Tabs` | `webtyp/widget/kind.go:16` | the Kind. `Kind.Allows` (kind.go:120) already admits `Selected` and `Current` for it. **No component currently returns it** |
| `widget.Current` | `webtyp/widget/state.go:45` | marks the active tab and the active panel |
| `style.SlideDeck(Motion)` | `webtyp/widget/style/flow.go:173` | stacks the panels as layers and reveals the one carrying `widget.Current`, sliding it in. Not a scroller |
| `BindStateFunc(widget.Current, fn)` | `webtyp/dom` | binds that state to a signal |

`layout/platformd` already does exactly this at page scale — read both before
writing a line:

```go
// layout/platformd/platformd.go:530 — the panel
panel := Section().Set(clsPanel.AsAttr()).
	Key(id).
	Attr("data-id", id).
	BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

// layout/platformd/platformd.go:556 — the nav link
link := A("#"+id).Set(clsNavLink.AsAttr()).
	Attr("data-id", id).
	BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })
```

```go
// layout/platformd/css.go:132 — the deck
Part(widget.Part("stage"),
	style.SlideDeck(style.MotionBase),
	style.Fill(),
).
Part(widget.Part("panel"),
	style.Stack(style.SpaceNone),
	style.Scroll(),
).
```

`tabs` is that mechanism shrunk from "the application shell" to "one region of
one screen".

## 3. Design gate

### Prior art

- **ARIA Authoring Practices — Tabs pattern.** A `tablist` of `tab` elements, a
  `tabpanel` per tab, `aria-selected` on one tab, `aria-controls` linking them.
  We follow its roles and attributes exactly; there is nothing to improve there
  and deviating costs screen-reader users.
- **Radix UI Tabs.** `Root` holds a `value`; `List`, `Trigger`, `Content` read
  it. Selection is one string, and every part derives from it. We follow that
  shape: one `*dom.SignalString` is the whole state.
- **Material Design tabs.** Adds a sliding indicator and scrollable tab strips
  for overflow. We take the first (it falls out of the existing style
  vocabulary) and **not** the second: a scrollable tab strip is a horizontal
  scroller, and `platformd`'s own comment records what horizontal scrollers cost
  here — a scroller inside a screen chains with the scroller a panel's content
  already has, and a swipe in the content drags the user to another section. A
  tab strip that does not fit wraps.

**Where we differ from all three:** nothing renders on a client framework's
state. Reveal is CSS, from `style.SlideDeck` reading a state attribute. All
panels stay mounted; the state decides which is on screen. That is the existing
`SlideDeck` contract, not a new one.

### Novice-name test

- "Tabs with these items" → `&tabs.Tabs{Items: []tabs.Item{...}}`.
- "This tab's label, icon, and panel" → `tabs.Item{ID, Label, Icon, Panel}`.
- "Which one is active" → `Active *dom.SignalString`.
- "Tell me when it changes" → `OnChange func(id string)`.

A junior reads `tabs.Tabs{Items: ..., Active: sig}` and can say what it does
without opening the file.

### Complexity ledger

| | change |
|---|---|
| Concepts | `+0` — `widget.Tabs`, `widget.Current` and `style.SlideDeck` all exist. This is their first component-scale consumer, not a new idea |
| Files | `+4` (`tabs/tabs.go`, `tabs/css.go`, `tabs/tabs_test.go`, `tabs/README.md`); 2 edited (`conformance_test.go`, `docs/CATALOG.md`) |
| Call-site lines | a consumer writes ~10 to get a tab set; hand-composing one costs ~90 plus its own stylesheet |
| Ways to do it | `−1` — a tabbed region had zero supported ways and was about to have one copied one |
| Net | negative |

### Where it belongs

`webtyp/components`, not `webtyp/layout`. `layout` holds page chassis —
`platformd` (the shell), `crudview` (a screen), `rightpanel` (a panel frame). A
tab set is a widget a screen contains, not a screen. It is also not an
application-level concern: the first consumer needs it, and the second one
always arrives.

### What it deletes

Nothing. It is the first component of its kind, and it prevents a deletion
debt rather than paying one.

## 4. Decisions already taken — do not revisit

1. **Panels are passed in, never built by this component.** `Item.Panel` is a
   `dom.Component`. `tabs` knows nothing about what a panel contains.
2. **All panels stay mounted.** That is `SlideDeck`'s contract, quoted in its
   doc: *"Todos los hijos siguen montados en el DOM. Ese es el trato: el estado
   decide cuál está en pantalla, nadie desmonta nada."* Do not add lazy
   mounting, and do not skip rendering inactive panels.
3. **No scrollable tab strip.** The strip wraps (`style.Row` + wrapping). See
   the Material note in §3.
4. **`Active` is injected, not owned.** The consumer passes a
   `*dom.SignalString` so the active tab can be read, set and persisted from
   outside (a URL fragment, a saved preference). A nil `Active` gets one
   created in `Init` defaulting to the first item's ID — convenience, not a
   second source of truth.
5. **Selection is a click handler, not a CSS `:checked` radio.** `SlideDeck`
   reveals the child carrying the `widget.Current` **state attribute**; CSS
   alone cannot move a state attribute. A radio-driven reveal would be a second,
   parallel reveal mechanism invented locally — exactly the leaf fix that is
   forbidden when the root already has the recipe.

**DEVIATION 5 — the component is `decktabs`/`DeckTabs`, not `tabs`/`Tabs`.**
`components/AGENTS.md` carries a rule the drafted plan broke: *"two words, and
the second word must name the class"* — a package and struct name must be at
least two words, and the pair must say **which style** of the thing it is. A
bare `tabs`/`Tabs` claims the whole concept for one implementation, leaving no
name for the day someone needs tabs that unmount their inactive panels.

`deck` is the characteristic, deliberately chosen over `slide`: the deck — every
panel mounted as a layer — is what this component *is*, while the slide is a
setting `MotionNone` removes without changing anything else. A name that
describes a setting is wrong as soon as someone changes it.

Everything below that says `tabs` / `Tabs` reads `decktabs` / `DeckTabs`:
package, folder, files (`decktabs.go`, `decktabs_test.go`), struct, widget name
(`widget.Name("decktabs")`), test prefixes, README and catalog entry.

This defect has the same root as DEVIATION 4: the rule lived in
`components/AGENTS.md`, and the **`components` skill did not carry it** — the
skill is what gets read when a component is created. Fixed in §8.

## 5. Stages

### Stage 1 — `decktabs/decktabs.go`

Package `tabs`. Follow `components/infobar/infobar.go` for the exact house
style: dot-imports of `webtyp.com/dom` and `webtyp.com/html`, a `widget.Name`
constant, `widget.Part` constants, and package-level `cls*` vars derived from
them. **Never write a class string by hand** — every class comes from
`NameTabs.Root()` or `NameTabs.Class(Part…)`.

```go
const NameTabs = widget.Name("tabs")

const (
	PartList  = widget.Part("list")
	PartTab   = widget.Part("tab")
	PartIcon  = widget.Part("icon")
	PartLabel = widget.Part("label")
	PartDeck  = widget.Part("deck")
	PartPanel = widget.Part("panel")
)
```

```go
// Item is one tab and the panel it reveals.
type Item struct {
	// ID identifies the tab. It is what Active carries and what the ARIA
	// attributes link by, so it must be unique within one Tabs and stable
	// across renders.
	ID    string
	Label string
	// Icon is optional; "" renders no icon.
	Icon svg.Icon
	// Panel is the content revealed when this tab is active. It stays mounted
	// whether or not it is on screen.
	Panel Component
}

// Tabs renders a strip of tabs over a deck of panels.
type Tabs struct {
	Element // value embed — never a pointer
	Items   []Item
	// Active carries the ID of the tab on screen. Optional: nil gets a signal
	// created in Init, defaulting to Items[0].ID.
	Active *SignalString
	// OnChange fires after Active changes. Optional.
	OnChange func(id string)
}

func (t *Tabs) WidgetName() widget.Name { return NameTabs }
func (t *Tabs) WidgetKind() widget.Kind { return widget.Tabs }
```

`Init(_ Ctx)` creates `Active` when nil and sets it to `Items[0].ID` when it is
empty. Guard `len(t.Items) == 0`.

`Render() *Element` builds two children of the root:

1. **The tab strip** — a `Nav` carrying `clsList`, `Attr("role", "tablist")`.
   One `Button` per item, carrying `clsTab` and:
   - `Attr("data-id", item.ID)`
   - `Attr("role", "tab")`, and `aria-controls` wired per DEVIATION 4
   - `BindStateFunc(widget.Current, func() bool { return t.Active.Get() == id })`
   - the icon via `item.Icon.Render(string(clsIcon))` when non-empty, then a
     `Span` with `clsLabel` and the label
   - **`type="button"`** — a bare `<button>` inside a form defaults to
     `type="submit"`, so a tab set placed in a form would submit it on every
     tab change.
2. **The deck** — a `Div` carrying `clsDeck`. One `Section` per item, carrying
   `clsPanel` and:
   - `Attr("data-id", item.ID)`
   - `Attr("role", "tabpanel")`, and `aria-labelledby` wired per DEVIATION 4
   - `BindStateFunc(widget.Current, func() bool { return t.Active.Get() == id })`
   - `.Child(item.Panel)` when `item.Panel != nil`

**DEVIATION 4 — `dom` mints the ids; the component never composes one.** The
drafted plan had the component write `id="tab-<Item.ID>"` / `id="panel-<Item.ID>"`
and reference those strings from `aria-controls` / `aria-labelledby`. That is
wrong, and not cosmetically:

- `dom` resolves every handler and every signal patch **by id** — an event is
  wired to `#3`, a signal patches `#3`. `dom/dom.go`'s "one id, one node"
  section spells out the failure: two nodes sharing an id give *"a component
  that renders, looks right, and does nothing, because the runtime resolved the
  other one"*.
- `dom`'s `claimID` **panics** when one render pass writes the same id twice.
  Two `Tabs` on one page fed the same `Item.ID`s — a staff screen and a patient
  screen both using `{"data","hours"}` — would have taken the render down.
- `Element.GetID()` already mints a unique id on demand, and `Element.For(other)`
  documents that mechanism as being *for label/input pairing and `aria-*`
  references*. `selectsearch` already wires `Attr("aria-controls", optList.GetID())`.

So both elements are built first and the pair is wired from their minted ids:

```go
tab.Attr("aria-controls", panel.GetID())
panel.Attr("aria-labelledby", tab.GetID())
```

`Item.ID` stays reachable as `data-id` — data for a test or a consumer query,
not an identifier the framework resolves.

`Element.Key()` was dropped too: it is the stable identity for keyed
reconciliation in `BindChildren`, and this component uses none, so setting it
was inert noise.

This defect came from the **`components` skill**, which said nothing about ids.
The skill was rewritten as part of this work (see §8).

**Anti-footgun — the loop variable.** Capture `item := item` (or bind the ID to
a local) before every closure, exactly as `platformd.go:528` does with `m := m`.
Without it every tab binds to the last item.

**DEVIATION 1 — `OnClick`, not `OnMount` + `Get`.** The drafted plan called for
an `OnMount()` that looked each tab up by id and attached a handler. `dom.Element`
already exposes `OnClick(func(Event))` directly, and that is the house idiom
(`calendarslider` binds its own controls with `OnChange` the same way). The
handler is attached on the element as it is built, so there is no second pass
over the DOM and no id lookup that can silently miss:

```go
tab := Button().Set(clsTab.AsAttr()).
	Key(id).
	Attr("type", "button").
	// … roles and aria …
	BindStateFunc(widget.Current, func() bool { return t.isActive(id) }).
	OnClick(func(Event) { t.activate(id) })
```

Selection moves through one unexported method so the no-op case lives in one
place:

```go
// activate moves the selection and notifies. Re-clicking the active tab is a
// no-op: it must not re-fire OnChange, which a consumer may use to reload.
func (t *Tabs) activate(id string) {
	if t.Active == nil || t.Active.Get() == id {
		return
	}
	t.Active.Set(id)
	if t.OnChange != nil {
		t.OnChange(id)
	}
}
```

No build tag on `tabs.go`.

### Stage 2 — `tabs/css.go`

```go
//go:build !wasm

package tabs
```

`func (t *Tabs) RenderCSS() *css.Stylesheet`, built with `style.For(t)`, in the
shape of `components/infobar/css.go`. Required rules:

- `Root(style.Stack(style.SpaceNone), style.Fill())` — strip above deck, deck
  takes the remaining height.
- `Part(PartList, style.Row(style.Space2), style.KeepSize(), style.DividerBelow(…))`
  — the strip does not shrink and is separated from the deck by a hairline.
- `Part(PartTab, style.Button(style.Subtle), style.Row(style.Space2), style.CenterContent())`.

  **DEVIATION 2 — `Button(Surface)`, not hand-composed `Interactive` +
  `ControlBox`.** `style.Button` is documented as *the one recipe for a button:
  the shared control height, inline padding for its label, and a box that
  neither stretches to its container nor shrinks under pressure*, and its doc
  states explicitly that the parts are **not safely composable by hand** —
  `KeepSize()` governs the main axis while a button in a `Stack` is stretched
  across the cross axis, and a component that composed its own button once
  shipped an 800px-wide bar. A tab is pressed, so it is a button. `Subtle` keeps
  the resting tab quiet. No separate hover rule: `Button` already derives hover,
  focus and press from its own family, and adding one is redundant
  composition.
- `When(widget.Current, PartTab, …)` — the active tab's resting paint. Copy the
  call shape from `layout/platformd/css.go:296`, which uses
  `When(widget.Current, widget.Part("nav-link"), …)`.
- `Part(PartDeck, style.SlideDeck(style.MotionBase), style.Fill())`
- `Part(PartPanel, style.Stack(style.SpaceNone), style.Scroll())`
- `Part(PartIcon, style.IconBox(style.IconSm), style.KeepSize())`
- `Part(PartLabel, style.FontSize(style.TextSm))`

**Hard rules, each of which the repository's own conformance test enforces:**

- **No `Anchor()` on the panel.** `platformd/css.go:138` records why: `SlideDeck`
  already positions each layer absolutely, which makes it the containing block
  for its content. `Anchor()` emits `position:relative` in `@layer widgets`,
  which beats the `position:absolute` the flow emits in `@layer primitives`, and
  the layers fall back into normal flow stacked one below the other.
- **`Scroll()` on the panel, not `Fill()`.** A panel taller than its layer must
  scroll inside that layer. `Scroll()` is `Fill()` plus `overflow-y`.
- **No `:root` block and no `RootCSS()`.** Theme tokens are the application's.
- **No hardcoded colors, sizes or spacing.** Every value comes from a `style`
  recipe or an allowed token. `conformance_test.go` holds the allowed variable
  list and will fail on anything else.
- **If a rule you need has no `style` recipe, stop.** That is a defect in
  `webtyp/widget/style`, and it gets its own plan in that repository — it is
  never hand-composed here. Report it instead of working around it.

### Stage 3 — `tabs/tabs_test.go`

No build tag (backend by default), following
`components/contentcard/card_test.go`. Required cases:

| Test | Asserts |
|---|---|
| `TestTabs_RendersOneTabAndPanelPerItem` | three items → three `role="tab"` and three `role="tabpanel"` |
| `TestTabs_FirstItemActiveByDefault` | with `Active` nil, the first tab and first panel carry the `Current` state attribute and no other does |
| `TestTabs_RespectsInjectedActive` | `Active` preset to the second ID → the second pair carries it |
| `TestTabs_AriaLinksTabToPanel` | every `aria-controls` matches an existing panel `id`, and every `aria-labelledby` an existing tab `id` |
| `TestTabs_TabsAreTypeButton` | no rendered tab lacks `type="button"` |
| `TestTabs_AllPanelsStayMounted` | an inactive item's panel content is present in the HTML |
| `TestTabs_EmptyItemsRendersNothingAndDoesNotPanic` | `Items: nil` → `Render()` returns without panicking |
| `TestTabs_RenderCSSNotNil` | `RenderCSS()` returns a non-nil stylesheet |
| `TestTabs_AriaLabelOnlyWhenSet` | **DEVIATION 3** — an optional `Label` field was added and becomes the strip's `aria-label`; absent when unset |
| `TestTabs_ActivateIsANoOpOnTheActiveTab` | re-selecting the tab already on view does not re-fire `OnChange` |
| `TestTabs_WidgetKindIsTabs` | `WidgetKind()` returns `widget.Tabs` |

**DEVIATION 3 — an optional `Label`.** The drafted plan left the tab strip
unnamed. ARIA wants a `tablist` labelled, and the consumer is the only one who
knows the name of the set, so `Label string` was added and is emitted as
`aria-label` only when non-empty. It needs no translation: the consumer supplies
it already in the reader's language, like every other label here.

Use `widget.Current.Key()` / `.Value()` to assert the state attribute rather
than writing the attribute string by hand — the state's own encoding is the
source of truth.

### Stage 4 — register in the root conformance test

`conformance_test.go` enumerates every component **twice**, and a component
missing from either list is silently unverified.

1. Add the import `"webtyp.com/components/tabs"` to the import block (keep it
   alphabetical: after `statgrid`, before `targethour`).
2. Add `&tabs.Tabs{},` to the `components := []interface{ RenderCSS() *css.Stylesheet }{...}`
   slice at line ~325, in alphabetical position.
3. Add `"tabs": &tabs.Tabs{},` to the `packageComponents` map in
   `TestKindAllowsEveryState` at line ~358.

`TestKindAllowsEveryState` is what proves `widget.Tabs` admits the states this
component uses. It is the reason step 3 is not optional.

### Stage 5 — documentation

- New `tabs/README.md` following `components/infobar/README.md`: what it is, the
  `Tabs` and `Item` structs, a wiring example with an injected `Active`, and the
  three decisions a consumer can get wrong (panels stay mounted; the strip
  wraps and never scrolls; `Active` is theirs to own).
- Add a `tabs` row to `docs/CATALOG.md`.
- Do not touch `docs/PLAN.md`, and do not link any permanent document to a plan
  file.

## 6. Stages table

| # | Stage | Files | Acceptance |
|---|---|---|---|
| 1 | Component | `tabs/tabs.go` (new) | renders strip + deck, ARIA linked, `Current` bound |
| 2 | Stylesheet | `tabs/css.go` (new) | `SlideDeck` deck, `Scroll` panel, no `Anchor` |
| 3 | Tests | `tabs/tabs_test.go` (new) | 8 cases green |
| 4 | Conformance | `conformance_test.go` | present in **both** lists |
| 5 | Docs | `tabs/README.md` (new), `docs/CATALOG.md` | catalogued |

## 7. Acceptance criteria

- `grep -rn "style.Anchor()" tabs/` → **empty**. (The drafted criterion was
  `grep -rn "Anchor()"`, which is wrong: it also matches the `css.go` comment
  that explains *why* `Anchor()` must not be there. The criterion must match
  the call, not the word.)
- `grep -rn ":root" tabs/` → **empty**; no `RootCSS` anywhere in `tabs/`.
- `grep -rnE '"tabs-|"tabs_' tabs/*.go` → **empty**. Every class is derived from
  `NameTabs`.
- `grep -c "tabs" conformance_test.go` → at least 3 (import + both lists).
- `grep -rn "type=\"button\"" tabs/tabs.go` → present.
- `grep -rnE '#[0-9a-fA-F]{3,8}|[0-9]+px|[0-9]+rem' tabs/css.go` → **empty**.
- `WidgetKind()` returns `widget.Tabs`.
- No stdlib import in `tabs/tabs.go` — `webtyp.com/fmt` and friends only.
- `gotest ./...` green across the whole repository, with no change to any other
  component.
- `docs/PLAN.md` unmodified.
- `grep -rnE 'Attr\("id"|"tab-"\+|"panel-"\+|\.Key\(' tabs/*.go` → **empty**.
  No id is composed by the component and no inert `Key` remains.
- `TestTabs_TwoSetsSharingItemIDsDoNotCollide` passes: two tab sets built from
  the same `Item.ID`s emit no duplicate id.

## 8. Fixed upstream: the `components` skill

DEVIATION 4 was a skill defect, not a one-off slip: the `components` skill
never mentioned ids. Auditing it against the repository showed it was stale in
almost everything structural — verified with a grep over `webtyp/components`:

| The skill said | Reality |
|---|---|
| CSS/SVG live in `ssr.go` | **no `ssr.go` exists**; the convention is `css.go` / `svg.go` |
| `//go:embed mycomponent.css` | **no `.css` file exists** anywhere |
| `RenderCSS() string` | `RenderCSS() *css.Stylesheet`, built with the `style` DSL |
| tokens `--mag-pri`, `--mag-sec`, `--mag-cua` | not in the allowed list (109 tokens: `--space-*`, `--color-*`, `--text-*`, …) |
| `dom.Div().Class("mycomponent")` | **no component uses a class literal**; classes derive from `widget.Name`/`widget.Part` |
| `c.Render().RenderHTML()` | `.String()`, with single-quoted attributes |
| `IconSvg() map[string]string` of raw SVG strings | `IconSvg() *sprite.Sprite`, built with `sprite.NewSprite`/`Define`/`Path` — all 11 real implementations |
| *(no mention of `webtyp/icons`)* | the shared glyph set, **one package per glyph**, which is what keeps each glyph's geometry behind its own `//go:build !wasm` |
| *(no mention of images)* | `image.go` + `RenderImages() []image.Asset`, read **by AST** so it must stay a hand-written literal, plus the `webtyp/image` builders |
| *(no mention of naming)* | the two-word rule from `components/AGENTS.md` — the rule DEVIATION 5 broke |

`webtyp/devskills/skills/components/SKILL.md` was rewritten against the
verified repository and rebuilt + synced (`go install ./cmd/devskills &&
devskills -f`) into all six agent configs. It now carries four sections that did not exist:

- **"Naming — two words, and the second must name the generic class"**, with
  the rule, why a bare generic name is unrecoverable, the "pick the
  characteristic that survives its own settings" test, and both times the rule
  was violated in this repo.
- **"Element IDs are MINTED by `dom` — never composed by a component"** — the
  rule, the two failure modes, the `GetID()` pattern, the `Key`-vs-id
  distinction, and a test rule: never assert on an element id you composed.
- **"Icons — `webtyp/icons` glyphs + `svg.go`"** — the reference/geometry split,
  the shared per-glyph packages, the real `*sprite.Sprite` signature.
- **"Images — `image.go` and the `webtyp/image` builders"** — the filename
  contract, the AST-literal constraint, the variant-bitmask waste rule, and the
  render builders.
