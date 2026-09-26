---
PLAN: "feat(components): inboxlist, bubblethread, composebar, presencelist — the pieces of a chat screen"
TAG: v0.7.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Stage F2a of** `veltylabs/mjosefa-cms/docs/MASTER.md` (private — context restated here). The
> consumer is the layout `webtyp.com/layout/chatview` (stage F2b), used by the module
> `github.com/veltylabs/chat_room`.

# Plan — four chat components

## 0. Context

A chat screen needs four pieces this repo does not have. Checked on 2026-09-26:

- `targetlist` renders only `ID`/`Label`/`Description` of a `view.Item` and is built around the
  CRUD selection/delete mode (`listselect`). It has no unread count and no time column. A
  conversation list is a different list, not a mode of that one.
- Nothing renders a message thread, a send box, or a list of people with an online dot.
- `countbadge` exists and is reused as-is for the unread bubble.
- `dom` has typed keys (`dom.KeyEvent`, `dom.KeyEnter`) but `KeyEvent` does **not** expose Shift.
  So the send box is **single-line**: Enter sends. Do not add modifier keys to `dom` in this plan.

The data arrives already shaped for display: times as `"HH:MM"` text, the preview already cut to
80 characters by the server. **No component formats dates or truncates text.**

## 1. Rules (restated from the skills `components`, `dom-elements`, `widget-styling`)

- Package and struct names are the ones below, exactly (two words: characteristic + generic
  class).
- Files per component: `<name>.go`, `css.go` (`//go:build !wasm`), `<name>_test.go`
  (`//go:build !wasm`), `<name>_wasm_test.go` (`//go:build wasm`) when behaviour needs the
  browser, `README.md`. No `svg.go`: none of these ship glyphs.
- Identity: `const Name<X> = widget.Name("<name>")`, one `widget.Part` per styled element,
  `cls… = Name<X>.Class(Part…)`, applied with `.Set(cls.AsAttr())`. **Never** `.Class("literal")`.
- Dot-import `webtyp.com/dom` and `webtyp.com/html`. Ids are minted by `dom`; never compose one.
  Events through the typed methods (`OnClick`, `OnKeyDown(func(KeyEvent))`, `OnInput`).
- Styles only through `webtyp.com/widget/style` (`style.For(c).Part(…).Stylesheet()`), tokens
  only (`style.Space*`, `style.Text*`, `style.Radius*`, `style.As(...)`). **If a needed recipe does
  not exist in `widget/style`, stop and report it in the PR description.** Do not hand-compose CSS
  and do not write raw property strings.
- WASM-compiled code: `webtyp.com/fmt` instead of `fmt`/`strings`/`strconv`/`errors`; no
  `map[K]V`; no `reflect`. `_test.go` files tagged `!wasm` may use stdlib.
- Run `gotest` (native + browser in one command). **Both lanes must pass.**
- No `TODO`, no commented-out code.

## 2. Design gate

**1. Prior art.** Slack, WhatsApp Web and Matrix/Element all split a chat screen into the same
four pieces: a conversation list (title, last message preview, time, unread bubble), a thread of
bubbles (own messages to the end side, others to the start side, author on others' messages,
"read" under own), a send box, and a people list with presence dots. Each here is its own
component so a different layout (a support widget, a notification drawer) can reuse one without
the others.

**2. Novice-name test.** `InboxList` (a list of inboxes; sibling: `targetlist`, the CRUD list),
`BubbleThread` (a thread drawn as bubbles; sibling: a compact IRC-style `LineThread`),
`ComposeBar` (a bar to compose a message; sibling: a multi-line `ComposePanel`), `PresenceList`
(a list of people with presence; sibling: a `RosterGrid` of avatars). Methods: `SetRows`,
`SetBubbles`, `Append`, `SetPeople`.

**3. Complexity ledger.**
```
Concepts        +4 components, +4 row types (Row, Bubble, Person; ComposeBar has none) 
Ways to do the same thing   0 (none of these exist)
```

**4. Where it belongs.** Here: they are raw pieces with no layout knowledge. The layout that
arranges them is `webtyp/layout/chatview` (another repo, next stage).

**5. What it deletes.** Nothing; new capability. Also registers the already-existing `countbadge`
in `conformance_test.go`, where it is missing today.

## 3. `inboxlist` — `InboxList` (`widget.Listbox`)

```go
type Row struct {
	ID      string
	Title   string
	Preview string // already short; rendered as-is
	Time    string // "HH:MM" or a short date; rendered as-is
	Unread  int
}

type InboxList struct {
	Element
	Selected *SignalString    // optional; created when nil. Holds the selected Row.ID.
	OnSelect func(id string)  // called when a row is clicked; also sets Selected
	Empty    string           // text shown when there are no rows; "" shows nothing

	// unexported: items []Row, rows *SignalNodes
}

func (l *InboxList) SetRows(rows []Row)
func (l *InboxList) Rows() []Row
```

- Parts: `PartList`, `PartRow`, `PartTitle`, `PartTime`, `PartPreview`, `PartEmpty`.
- Each row is a `Button` (type `button`, an `Anchor` host for the bubble) containing title and
  time on one line, preview below, and a `countbadge.CountBadge` with `Count` = `fmt.Sprint(Unread)`
  and `Visible` = `Unread > 0`. The row carries `widget.Selected` via `BindStateFunc` when
  `Selected.Get() == row.ID`, and `aria-selected` accordingly.
- `SetRows` replaces the list, **keeping** the `Selected` value; order is exactly the order given.
- A row with `Unread > 0` has its title in bold (`style.FontWeight(style.WeightBold)` on a
  dedicated part `PartTitleUnread`; choose the part when rendering, no state misuse).

## 4. `bubblethread` — `BubbleThread` (`widget.Region`, `role="log"`, `aria-live="polite"`)

```go
type Bubble struct {
	ID     string
	Author string // shown only when Mine is false
	Body   string
	Time   string
	Mine   bool
	Read   bool   // meaningful only when Mine is true
}

type BubbleThread struct {
	Element
	ReadLabel string // text under a Mine && Read bubble, e.g. "Leído"; "" renders nothing
	Empty     string // text shown with no bubbles; "" renders nothing
}

func (t *BubbleThread) SetBubbles(b []Bubble)  // replace everything
func (t *BubbleThread) Append(b ...Bubble)     // add at the end, skipping IDs already present
func (t *BubbleThread) MarkRead(ids ...string) // set Read on those bubbles, re-render them
func (t *BubbleThread) Bubbles() []Bubble
```

- Parts: `PartThread` (scrolling container, `style.Scroll()`), `PartMine`, `PartTheirs`,
  `PartAuthor`, `PartBody`, `PartTime`, `PartRead`, `PartEmpty`.
- `PartMine` is pushed to the end side (`style.PushEnd()` within a `style.Stack`), `PartTheirs`
  stays at the start. Surfaces: `PartMine` → `style.As(style.AccentWash)` (what `targetlist`
  uses for a selected row), `PartTheirs` → `style.As(style.Inset)` (what `contentcard` uses for
  its body), so the chat speaks the same visual language.
- Body text is inserted as **text** (`Text(...)`), never as HTML.
- After `SetBubbles` and after an `Append` that added at least one bubble, the last bubble is
  scrolled into view (`dom.Get(id)` on the last bubble's minted id, then
  `ScrollIntoViewInstant()`). On the backend lane this is a no-op.

## 5. `composebar` — `ComposeBar` (`widget.Form`)

```go
type ComposeBar struct {
	Element
	Placeholder string
	SendLabel   string            // button text, e.g. "Enviar"
	MaxLength   int               // > 0 sets maxlength on the input; <= 0 sets none
	OnSend      func(body string) // called with the trimmed, non-empty text
	Disabled    *SignalBool       // optional; nil = always enabled
}
```

- Parts: `PartBar`, `PartInput`, `PartSend`.
- One `<input type="text">` two-way bound (`Bind`) to an internal `*SignalString`, and one
  `<button type="button">` with `SendLabel`.
- Send happens on button click **or** on `OnKeyDown` with `e.Key() == KeyEnter` (call
  `e.PreventDefault()`). Send = trim the text with `fmt.TrimSpace(s)` (`webtyp.com/fmt`); if empty, do nothing; else call `OnSend(text)` and clear the input.
- `OnSend == nil` panics at `Render` with `"composebar: OnSend is required"` — a send box that
  cannot send is a bug, not a configuration.
- When `Disabled` is true, both input and button carry `disabled` and `widget.Disabled`.

## 6. `presencelist` — `PresenceList` (`widget.Listbox`)

```go
type Person struct {
	ID     string
	Label  string
	Online bool
}

type PresenceList struct {
	Element
	OnSelect     func(id string) // row clicked
	OnlineLabel  string          // screen-reader text for the online dot, e.g. "En línea"
	OfflineLabel string          // e.g. "Desconectado"
	Empty        string
}

func (l *PresenceList) SetPeople(p []Person) // renders online first, then offline; each group by Label (A→Z)
func (l *PresenceList) People() []Person
```

- Parts: `PartList`, `PartRow`, `PartDotOnline`, `PartDotOffline`, `PartLabel`, `PartStatus`,
  `PartEmpty`.
- Each row is a `Button` with the dot, the label, and a `PartStatus` span holding
  `OnlineLabel`/`OfflineLabel` styled with `style.VisuallyHidden()` (the dot's colour alone must
  not carry the meaning).
- Sorting uses a simple insertion sort over the slice — no `sort` import in WASM code, no map.

## 7. Registration and docs

- `conformance_test.go`: add the four imports alphabetically, add `&inboxlist.InboxList{}`,
  `&bubblethread.BubbleThread{}`, `&composebar.ComposeBar{OnSend: func(string) {}}`,
  `&presencelist.PresenceList{}` to the stylesheet slice **and** to `TestKindAllowsEveryState`'s
  map; also add `countbadge` to both (missing today).
- `docs/CATALOG.md`: one section per component, same format as the existing ones.
- Each `README.md`: import, usage example, fields.

## 8. Tests

Per component, `!wasm` lane:
1. `Render()` is idempotent (same string twice) and contains the root class.
2. `RenderCSS()` builds without panic (validates the stylesheet).
3. `InboxList`: `SetRows` with 2 rows → 2 row parts; `Unread: 0` → the badge is not open;
   `Unread: 3` → badge text `3` and open; `PartTitleUnread` only on the unread row.
4. `BubbleThread`: a `Mine` bubble has `PartMine` and no author; a theirs bubble has
   `PartTheirs` and the author; `ReadLabel` appears only under `Mine && Read`; `Append` with an
   existing ID does not duplicate; body `<b>x</b>` is escaped in the output.
5. `ComposeBar`: `Render` with `OnSend == nil` panics with the exact message; `MaxLength: 2000`
   → `maxlength='2000'`.
6. `PresenceList`: `SetPeople` of `[{b offline}, {a online}, {c online}]` renders `a, c, b`.

Browser lane (`_wasm_test.go`):
7. `InboxList`: clicking a row calls `OnSelect` with its ID and sets `Selected`.
8. `ComposeBar`: typing `"  hola "` and pressing Enter calls `OnSend("hola")` once and empties
   the input; Enter on an empty input calls nothing; the button does the same as Enter.
9. `BubbleThread`: after `Append`, the last bubble is in the DOM (scrolling itself is not
   asserted).
10. `PresenceList`: clicking a row calls `OnSelect`.

## 9. Stages

| # | Stage | Files | Acceptance |
|---|---|---|---|
| C1 | `inboxlist` | `inboxlist/*` | tests 1–3, 7 green |
| C2 | `bubblethread` | `bubblethread/*` | tests 1, 2, 4, 9 green |
| C3 | `composebar` | `composebar/*` | tests 1, 2, 5, 8 green |
| C4 | `presencelist` | `presencelist/*` | tests 1, 2, 6, 10 green |
| C5 | Registration + docs | `conformance_test.go`, `docs/CATALOG.md`, READMEs | conformance suite green; `grep -rn '\.Class("' inboxlist bubblethread composebar presencelist` empty; `grep -rn "map\[\|TODO" inboxlist bubblethread composebar presencelist --include=*.go` empty; `gotest` green on both lanes |
