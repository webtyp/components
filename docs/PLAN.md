---
PLAN: "feat(components): stepaccordion (sequential wizard accordion) and chipgrid (single-choice chip grid)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 13116045847803051040
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

> **Phase A** of
> [`veltylabs/mjosefa-cms/docs/RESERVA_Y_PAGOS_MASTER_PLAN.md`](https://github.com/veltylabs/mjosefa-cms/blob/main/docs/RESERVA_Y_PAGOS_MASTER_PLAN.md)
> (decision RP-1). Gate for Phase B (`veltylabs/appointment_booking`), which builds a
> 6-step booking flow out of these two components.

# Plan — two new components: `stepaccordion` and `chipgrid`

## 0. Context (read first)

A booking flow needs (1) a **vertical wizard**: steps open one at a time; a finished step
collapses to a one-line summary with an "Edit" button; a step below the open one is locked
until every step above it is finished; and (2) a **grid of chips** where the user picks one
value (a free hour such as `09:30`).

Neither exists in this repo. The closest pieces and why they do not fit:

| Existing | Why it is not the answer |
|---|---|
| `stepindicator` | Horizontal progress bar; it does not hold each step's content nor collapse a finished step to its summary |
| `segmentedcontrol` | A pill switch for a handful of options; 20 hours on a row do not fit, and it has no disabled member |
| `targethour` | A list of rows (hour + label + status), built for an agenda, not for picking one hour |
| `cellgrid` | Spreadsheet-style paint grid with row/column coordinates, not labelled single choice |

A previous attempt built both pieces inside the consuming module with hand-written class
strings and no stylesheet; it shipped unstyled. That is why they live here, with identity,
`css.go` and conformance like every other component.

## Design gate

### 1. Prior art
- **Wizard accordion**: Doctolib and Zocdoc booking flows (each answered question collapses
  to a summary with "Change"); GOV.UK "check your answers" summary rows with a "Change" link;
  Material 3 vertical stepper; WAI-ARIA Accordion pattern (`button` with `aria-expanded`
  + `aria-controls`, the panel `role="region"` with `aria-labelledby`).
- **Chip grid**: Calendly / Doctolib time-slot grids; Material 3 "choice chips"; WAI-ARIA
  Listbox pattern (`role="listbox"`, `role="option"`, `aria-selected`, roving `tabindex`,
  arrow keys).

### 2. Novice-name test
- `stepaccordion.StepAccordion` — "an accordion whose sections are steps done in order".
  Coexisting sibling: a free accordion whose sections open in any order
  (`freeaccordion.FreeAccordion`) — the name leaves room for it.
- `Step{Key, Title, Body}` — one step: its id, its heading, its content.
- `(*StepAccordion).Advance(summary)` — "finish the open step with this summary and open
  the next one".
- `(*StepAccordion).Reopen(key)` — "open that finished step again".
- `OnReopen func(key string)` — "tell me a step was reopened" (the host clears what depends
  on it).
- `chipgrid.ChipGrid` — "a grid of chips". Coexisting sibling: `cellgrid.CellGrid`
  (spreadsheet cells) already exists; a multi-choice variant would be
  `toggleset.ToggleSet` or similar.
- `Chip{Value, Label, Disabled}` — one choice.
- `(*ChipGrid).SetChips(chips)` — replace the choices.
- `OnPick func(value string)` — "tell me what was picked".

### 3. Complexity ledger
- New: 2 packages, 2 structs each with 3–4 exported fields, 2 small structs (`Step`, `Chip`),
  3 exported methods, 2 callbacks, 9 `widget.Part` constants.
- No new `widget.State`, no new `style` recipe, no change to any other package.
- Removed: nothing in this repo; in `appointment_booking` (Phase B) ~650 lines of the broken
  `booking_accordion.go` disappear because these exist.

### 4. Where it belongs
`webtyp/components`, one package per component, like every other reusable widget. The
consumer is a domain module UI (`veltylabs/*/ui`), which may import `webtyp.com/components/*`
but must not hand-compose styles.

### 5. What it deletes
Nothing here. Downstream it deletes `appointment_booking/ui/booking_accordion.go` (Phase B).

## 1. Repository rules the executor must follow (inline, no exceptions)

- **WASM/TinyGo first.** Non-test code imports only `webtyp.com/*` packages (`dom`, `html`,
  `fmt`, `lang`, `widget`, `widget/style`, `css`). **Never** `strings`, `strconv`, `errors`,
  stdlib `fmt`. Use `webtyp.com/fmt`.
- **No `map[K]V` anywhere**, tests included. Use slices + linear scan.
- **Value embed only**: `dom.Element` embedded as a value, never `*dom.Element`.
- **Identity**: declare `const NameX = widget.Name("…")` and `widget.Part` constants; derive
  every class with `NameX.Root()` / `NameX.Class(Part)` and apply with `.Set(cls.AsAttr())`.
  **`.Class("literal")` is forbidden.**
- **Never compose an element id.** Use `.Key(...)` when a stable identity is needed and let
  `dom` mint ids; to point `aria-controls`/`aria-labelledby` at another element, create that
  element first and read its minted id with `el.GetID()` — exactly as
  `selectsearch/selectsearch.go` does (`Attr("aria-controls", optList.GetID())`).
- **Stylesheet only in `css.go`** with `//go:build !wasm`, built with
  `style.For(c)….Stylesheet()`. Compose **only existing recipes** of `webtyp.com/widget/style`
  (`Stack`, `Row`, `Grid`, `ChipBox`, `Button`, `As`, `Pad`, `Round`, `Raise`, `FontSize`,
  `FontWeight`, `Divider`, `DividerBetween`, `Hide`, `RevealedBy`, …). **If a needed look
  is missing, stop and report it in the PR description — do not hand-write CSS.**
- **UI text is `lang.Text` in English**, translated in `Render` with
  `lang.Translate(x).String()`; every new English key is added to the root `lang.json` with
  its Spanish value (`tests/lang_json_test.go` enforces this).
- **Tests live ONLY in `tests/`** (external package `tests`, exported API only). Older
  components keep tests inside their folders; **these two do not**: no `_test.go` inside
  `stepaccordion/` or `chipgrid/`. Per component: `tests/<name>_test.go` (`//go:build !wasm`:
  render + `RenderCSS` + state logic) and `tests/<name>_ui_wasm_test.go` (`//go:build wasm`:
  mount in the DOM, click, type keys). Use `segmentedcontrol/` and `stepindicator/` tests as
  examples of style only. Run with `gotest` (never `go test`).

## 1.1 Method: diagram → failing tests → code (TDD), 100% coverage of new code

- Prerequisite: `go install webtyp.com/devflow/cmd/gotest@latest`.
- Source of truth: [`docs/diagrams/stepaccordion_chipgrid.md`](diagrams/stepaccordion_chipgrid.md)
  (already written). Every branch has its exact test name in §5.
- In stages 1 and 2: write that stage's tests from §5 first (create the exported signatures
  with `panic("not implemented")` so they compile) → `gotest` → **confirm they fail** → commit
  `test(<component>): failing tests` → implement → green → commit `feat(<component>): …`.
  The PR history must show each red commit before its green one.
- **Coverage: 100% of the statements of both new packages** in the non-wasm run:

  ```bash
  go test -coverpkg=webtyp.com/components/stepaccordion,webtyp.com/components/chipgrid \
    -coverprofile=/tmp/cover.out ./tests/... && go tool cover -func=/tmp/cover.out | grep -E "stepaccordion|chipgrid"
  # every function line must show 100.0%
  ```

  Click and key handlers only run in wasm, where coverage is not measured; for them the
  criterion is that every `P*`/`K*` branch of the diagram has its wasm test in §5. If a
  non-wasm line cannot be reached through the exported API, delete it; never export
  something to reach it.
- **Conformance**: register both components in `conformance_test.go` (import; the
  `components := []interface{…}` slice; the `TestKindAllowsEveryState` map).

## 2. Stage 1 — `stepaccordion`

Files to create:

```
stepaccordion/
├── stepaccordion.go
├── css.go                       //go:build !wasm
└── README.md
tests/
├── stepaccordion_test.go        //go:build !wasm
└── stepaccordion_ui_wasm_test.go //go:build wasm
```

### 2.1 API (exact)

```go
package stepaccordion

const NameStepAccordion = widget.Name("stepaccordion")

const (
	PartStep    = widget.Part("step")    // <section> per step
	PartHeader  = widget.Part("header")  // row: number badge + title + summary + edit button
	PartBadge   = widget.Part("badge")   // step number "1", "2", …
	PartTitle   = widget.Part("title")
	PartSummary = widget.Part("summary") // one line, shown only on a finished step
	PartEdit    = widget.Part("edit")    // "Edit" button, shown only on a finished step
	PartBody    = widget.Part("body")    // the step content, shown only on the open step
)

// Step is one section of the wizard.
type Step struct {
	Key   string        // stable, unique within the accordion
	Title lang.Text     // English; translated in Render
	Body  dom.Component // rendered inside PartBody while the step is open
}

// StepAccordion opens its steps one at a time, in order.
type StepAccordion struct {
	dom.Element
	Steps []Step
	// OnReopen runs after Reopen(key), so the host can clear the choices that
	// depended on that step. Optional.
	OnReopen func(key string)
	// unexported state (see 2.2)
}

func (a *StepAccordion) WidgetName() widget.Name { return NameStepAccordion }
func (a *StepAccordion) WidgetKind() widget.Kind { return widget.Disclosure }

func (a *StepAccordion) Init(_ dom.Ctx)
func (a *StepAccordion) Render() *dom.Element

// Advance marks the open step finished, shows summary on its header, and opens
// the next step. On the last step it marks it finished and opens nothing.
func (a *StepAccordion) Advance(summary string)

// Reopen opens the finished step key again. Every step after it goes back to
// locked and loses its summary. Then OnReopen(key) runs. Unknown key or a step
// that is not finished: no-op.
func (a *StepAccordion) Reopen(key string)

// Open returns the key of the open step, or "" when every step is finished.
func (a *StepAccordion) Open() string
```

### 2.2 State model

Each step is in exactly one of three phases: **locked**, **open**, **done**. Keep it as an
unexported slice of per-step records (no map), each holding:
`open *dom.SignalBool`, `done *dom.SignalBool`, `summary *dom.SignalString`.
Initial state after `Init`: step 0 open, the rest locked. Zero steps: render an empty root.

### 2.3 Markup (exact)

```
<div class=stepaccordion>                                   (Root)
  <section class=stepaccordion__step data-open? data-disabled?>  (one per step; Key(step.Key))
    <div class=stepaccordion__header>
      <span class=stepaccordion__badge>N</span>             (1-based)
      <h3 class=stepaccordion__title>Translated title</h3>
      <span class=stepaccordion__summary>…</span>           (BindText(summary); hidden unless done)
      <button type=button class=stepaccordion__edit>Edit</button>  (hidden unless done; OnClick → Reopen(key))
    </div>
    <div class=stepaccordion__body role=region>…Body…</div> (hidden unless open)
  </section>
</div>
```

- Bind the step's `open` signal to `widget.Open` and its locked state (a `DeriveBool` that is
  true when neither open nor done) to `widget.Disabled`, both with `BindState` on the
  `<section>`.
- Show/hide summary, edit and body with the `widget/style` reveal recipes keyed on those
  states (`RevealedBy`/`Hide`), never with inline `style`.
- The header of a locked step is not focusable and not clickable. Only the Edit button
  of a done step is interactive; the header itself has no click handler.
- `aria-expanded` on the Edit button: `"false"` (it only ever opens); its `aria-controls` is
  the body's `GetID()`. The body carries `aria-labelledby` = the title's `GetID()`.

### 2.4 Stylesheet

`Root`: vertical stack with a gap. `PartStep`: a panel surface, rounded, padded.
`PartHeader`: a row, items centred. `PartBadge`: a chip-sized round box. `PartSummary`:
small text, pushed to the end. `PartEdit`: a bare button. Locked step: dimmed via the
`Disabled` state recipe. Use only recipes listed in §1.

### 2.5 Text

English keys: `"Edit"`. Add `"Edit": ["Editar"]` to `lang.json` if absent.

### 2.6 Tests

See §5 (rows of stage 1).

## 3. Stage 2 — `chipgrid`

Files to create:

```
chipgrid/
├── chipgrid.go
├── css.go                    //go:build !wasm
└── README.md
tests/
├── chipgrid_test.go          //go:build !wasm
└── chipgrid_ui_wasm_test.go  //go:build wasm
```

### 3.1 API (exact)

```go
package chipgrid

const NameChipGrid = widget.Name("chipgrid")

const (
	PartChip  = widget.Part("chip")
	PartEmpty = widget.Part("empty") // shown when there are no chips
)

// Chip is one choice. Label is shown as is (it is data, e.g. "09:30", not UI text).
type Chip struct {
	Value    string
	Label    string
	Disabled bool
}

// ChipGrid shows its chips in a grid and lets the user pick one.
type ChipGrid struct {
	dom.Element
	// Selected holds the picked value, or "". Optional: created if nil so a host
	// can share it.
	Selected *dom.SignalString
	// Empty is the text shown when there are no chips. English; translated in Render.
	// Zero value: "No options available".
	Empty lang.Text
	// OnPick runs when the user picks an enabled chip. Optional.
	OnPick func(value string)
	// unexported: chips *dom.SignalNodes, the current []Chip
}

func (g *ChipGrid) WidgetName() widget.Name { return NameChipGrid }
func (g *ChipGrid) WidgetKind() widget.Kind { return widget.Listbox }

func (g *ChipGrid) Init(_ dom.Ctx)
func (g *ChipGrid) Render() *dom.Element

// SetChips replaces the choices. If Selected is no longer among the enabled
// chips it is cleared to "".
func (g *ChipGrid) SetChips(chips []Chip)
```

### 3.2 Markup and behaviour

```
<div class=chipgrid role=listbox>                (Root; BindChildren(chips))
  <button type=button class=chipgrid__chip role=option aria-selected=… data-selected? data-disabled?>09:00</button>
  …
  <p class=chipgrid__empty>No options available</p>   (only when there are no chips)
</div>
```

- Each chip binds `widget.Selected` to `DeriveBool(Selected == value)` and `widget.Disabled`
  to its `Disabled` flag; a disabled chip also carries the `disabled` attribute and ignores
  clicks.
- Click on an enabled chip: `Selected.Set(value)`, then `OnPick(value)`.
- Keyboard (WAI-ARIA listbox, roving tabindex): only the selected chip (or the first enabled
  one when none is selected) has `tabindex="0"`, the rest `"-1"`. `ArrowRight`/`ArrowDown` move
  focus to the next enabled chip, `ArrowLeft`/`ArrowUp` to the previous, `Home`/`End` to the
  first/last enabled one; `Enter`/`Space` pick the focused chip. Use the typed key events of
  `webtyp.com/dom` (`OnKeyDown`, `dom.KeyArrowRight`, …) as `segmentedcontrol` does. Copy the
  roving-tabindex mechanics from `calendarslider` (`days`/`dayRef`), which already implements
  them.

### 3.3 Stylesheet

`Root`: `style.Grid` with chip-sized columns that reflow by container width (the `Grid`
recipe's column cap). `PartChip`: `style.ChipBox` + `style.Button` look, rounded. Selected:
the primary/raised surface recipe. Disabled: the `Disabled` state recipe. `PartEmpty`:
muted small text, spanning the full row (`SpanFull`).

### 3.4 Text

English keys: `"No options available"` → `"No hay opciones disponibles"` in `lang.json`.

### 3.5 Tests

See §5 (rows of stage 2).

## 4. Stage 3 — docs and registration

- `README.md` in each package: one usage example (copy the shape of
  `segmentedcontrol/README.md`).
- Root `README.md` / `docs/CATALOG.md`: add one row per component next to the others.
- `conformance_test.go`: both components in the import block, the stylesheet slice and the
  `TestKindAllowsEveryState` map.

## 5. Tests — diagram → test traceability (all in `tests/`)

Read state attributes through `widget.Open.Attr()`, `widget.Disabled.Attr()`,
`widget.Selected.Attr()`; never hard-code them.

| Branch | Test | File | Stage |
|---|---|---|---|
| §1 initial phases | `TestStepAccordion_InitFirstOpenRestLocked` | `stepaccordion_test.go` | 1 |
| `R1-yes` | `TestStepAccordion_ZeroStepsRendersEmptyRoot` | `stepaccordion_test.go` | 1 |
| `R2-yes` | `TestStepAccordion_RenderWithoutInit` | `stepaccordion_test.go` | 1 |
| `A1-no` | `TestStepAccordion_AdvanceWithNothingOpenIsNoop` | `stepaccordion_test.go` | 1 |
| `A2-no` | `TestStepAccordion_AdvanceOpensNextAndSetsSummary` | `stepaccordion_test.go` | 1 |
| `A2-yes` | `TestStepAccordion_AdvanceOnLastLeavesNothingOpen` | `stepaccordion_test.go` | 1 |
| `B1-no` | `TestStepAccordion_ReopenUnknownKeyIsNoop` | `stepaccordion_test.go` | 1 |
| `B2-no` | `TestStepAccordion_ReopenLockedOrOpenIsNoop` | `stepaccordion_test.go` | 1 |
| `B3-yes` | `TestStepAccordion_ReopenLocksLaterStepsAndCallsOnReopen` | `stepaccordion_test.go` | 1 |
| `B3-no` | `TestStepAccordion_ReopenWithoutCallback` | `stepaccordion_test.go` | 1 |
| ARIA wiring | `TestStepAccordion_EditControlsBodyAndBodyLabelledByTitle` | `stepaccordion_test.go` | 1 |
| translated "Edit" | `TestStepAccordion_EditTextTranslated` | `stepaccordion_test.go` | 1 |
| stylesheet | `TestStepAccordion_RenderCSS` | `stepaccordion_test.go` | 1 |
| Edit click in the DOM | `TestWASM_StepAccordion_EditReopensInDOM` | `stepaccordion_ui_wasm_test.go` | 1 |
| locked header not interactive | `TestWASM_StepAccordion_LockedStepHasNoFocusableControl` | `stepaccordion_ui_wasm_test.go` | 1 |
| `Z1-yes` | `TestChipGrid_EmptyShowsDefaultText`, `TestChipGrid_EmptyShowsCustomText` | `chipgrid_test.go` | 2 |
| `Z1-no` | `TestChipGrid_RendersOneOptionPerChip` | `chipgrid_test.go` | 2 |
| `Z2-yes` | `TestChipGrid_SetChipsKeepsEnabledSelection` | `chipgrid_test.go` | 2 |
| `Z2-no` | `TestChipGrid_SetChipsClearsMissingOrDisabledSelection` | `chipgrid_test.go` | 2 |
| `Z3` both | `TestChipGrid_TabStopOnSelectedOrFirstEnabled` | `chipgrid_test.go` | 2 |
| shared `Selected` | `TestChipGrid_UsesHostSelectedSignal` | `chipgrid_test.go` | 2 |
| stylesheet | `TestChipGrid_RenderCSS` | `chipgrid_test.go` | 2 |
| `P1-yes` | `TestWASM_ChipGrid_DisabledClickIgnored` | `chipgrid_ui_wasm_test.go` | 2 |
| `P1-no/P2-yes` | `TestWASM_ChipGrid_ClickPicksAndCallsOnPick` | `chipgrid_ui_wasm_test.go` | 2 |
| `P2-no` | `TestWASM_ChipGrid_ClickWithoutCallback` | `chipgrid_ui_wasm_test.go` | 2 |
| `K1-yes` | `TestWASM_ChipGrid_ArrowsSkipDisabled`, `TestWASM_ChipGrid_HomeEnd` | `chipgrid_ui_wasm_test.go` | 2 |
| `K1-no` | `TestWASM_ChipGrid_ArrowAtEdgeStays` | `chipgrid_ui_wasm_test.go` | 2 |
| Enter/Space | `TestWASM_ChipGrid_EnterAndSpacePick` | `chipgrid_ui_wasm_test.go` | 2 |

## 6. Acceptance criteria

- `gotest` green (vet, race, tests, wasm).
- The coverage command of §1.1 shows 100.0% for every function of both packages.
- Every test of §5 exists with its exact name in `tests/`;
  `find stepaccordion chipgrid -name "*_test.go"` → empty.
- The PR history shows a red `test(...)` commit before each `feat(...)`.
- `grep -rn '\.Class("' stepaccordion chipgrid` → empty.
- `grep -rn 'map\[' stepaccordion chipgrid` → empty.
- `grep -rnE '"(strings|strconv|errors|fmt)"' stepaccordion/*.go chipgrid/*.go | grep -v _test.go` → empty.
- `lang.json` contains `"Edit"` and `"No options available"`.
- Both packages appear three times in `conformance_test.go`.
- Before opening the PR: `git diff --stat main` lists every file of the stages table below,
  and no file is deleted that this plan does not order deleted.

## 7. Stages table

| Stage | Files |
|---|---|
| 1 | `stepaccordion/stepaccordion.go`, `stepaccordion/css.go`, `stepaccordion/README.md`, `tests/stepaccordion_test.go`, `tests/stepaccordion_ui_wasm_test.go` |
| 2 | `chipgrid/chipgrid.go`, `chipgrid/css.go`, `chipgrid/README.md`, `tests/chipgrid_test.go`, `tests/chipgrid_ui_wasm_test.go` |
| 3 | `conformance_test.go`, `lang.json`, `README.md`, `docs/CATALOG.md` |
