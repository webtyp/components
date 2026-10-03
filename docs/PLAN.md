---
PLAN: "feat(components): add stepindicator, segmentedcontrol, and cellgrid widgets"
TAG: v0.8.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 12526991816552546311
PR: https://github.com/webtyp/components/pull/30
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — Reusable UI Components: StepIndicator, SegmentedControl, and CellGrid

This plan implements three pure, generic UI components in `webtyp/components`:
1. `stepindicator`: A numbered wizard/stepper control (active, done with checkmark, pending, mobile responsive label collapse).
2. `segmentedcontrol`: A horizontal pill switch for 2 or more options (sunken track, raised active pill, WAI-ARIA tab/tablist semantics).
3. `cellgrid`: An interactive, fluid 2D coordinate grid canvas (Excel bijective column labels A..Z/AA.., 1-based row numbers, fluid 1:1 square cells, pointer paint engine with pointer capture and Bresenham interpolation, and sparse cell rendering).

---

## Design gate

### 1. Prior art
- **Stepper / Steps:** Ant Design (`<Steps>`), Material-UI (`<Stepper>`), and Shoelace (`<sl-stepper>`). All model steps as an ordered slice of items with numeric badges, checkmarks on completion, and auto-collapsing labels on narrow viewports.
- **Segmented Control:** Apple iOS Human Interface Guidelines (`UISegmentedControl`), Mantine (`SegmentedControl`), and Radix UI (`ToggleGroup`). Provides a sunken background container with a sliding or raised active pill.
- **Coordinate / Cell Grid:** React Grid Layout, Handsontable, and ag-Grid. A pure 2D matrix supporting fluid cell coordinates, cell hit-testing, and drag selection. `webtyp/components/cellgrid` provides a lightweight, CSS-grid-native matrix tailored for WebAssembly with zero external dependencies.

### 2. Novice-name test
- `stepindicator`: Immediately tells a developer it displays step progress. Follows the two-word naming rule.
- `segmentedcontrol`: Standard industry term for segmented pill selectors.
- `cellgrid`: Clearly describes a grid of discrete cells with coordinate tracking.

### 3. Complexity ledger
```
Concepts the developer must learn   +3 (StepIndicator, SegmentedControl, CellGrid)
Files they must touch to do X       +1 (import and instantiate the component)
Lines at the call site              -40 (replaces hand-rolled HTML/CSS loops)
Ways to do the same thing           0 (canonical components for these three patterns)
```

### 4. Where does it belong
In `webtyp/components`. Each is a pure, domain-agnostic UI widget with no business logic or dependencies on `room_layout` or external modules. Conforms to the standard conformance suite in `conformance_test.go`.

### 5. What does this change delete?
Deletes ad-hoc stepper, segmented pill, and grid implementations repeated across applications.

---

- **No standard library in WASM packages:** Use `webtyp/fmt` instead of stdlib `fmt`, `errors`, `strconv`, or `strings`; `webtyp/time` instead of `time`; and `webtyp/json` instead of `encoding/json`.
- **No `map` declarations in WASM code:** Avoid using `map` declarations in WASM code to prevent TinyGo binary bloat. Use structs or slices (`[]CellCoord`, `[]string`, `[]Step`, `[]Option`) with linear or binary search for small collections instead.
- **Value embedding only:** Embed `dom.Element` as a value (`Element dom.Element`), never as a pointer.
- **SSR split:** Stylesheets live in `css.go` with `func (w *Widget) RenderCSS() *css.Stylesheet` (built with `//go:build !wasm`).
- **No hardcoded strings:** Export typed widget names:
  - `const NameStepIndicator = widget.Name("stepindicator")`
  - `const NameSegmentedControl = widget.Name("segmentedcontrol")`
  - `const NameCellGrid = widget.Name("cellgrid")`

---

## Stages

### Stage 1: `stepindicator` component
Create package `webtyp/components/stepindicator`:
- `stepindicator.go`:
  - Struct `StepIndicator`:
    - `Steps []Step` where `type Step struct { Key, Label string }`
    - `Active int` (0-based active step index)
    - `OnChange func(index int)`
  - Renders `<nav class="stepindicator">` with numbered circle badges, connecting lines, and responsive label suppression below 760px.
- `css.go`: Stylesheet adhering to WebTyp tokens (`--color-primary`, `--color-outline`, `--color-muted`).
- `stepindicator_test.go`: DOM structure tests.
- `stepindicator_ui_wasm_test.go`: Interactive click and state change tests.
- `README.md`: Usage documentation.

### Stage 2: `segmentedcontrol` component
Create package `webtyp/components/segmentedcontrol`:
- `segmentedcontrol.go`:
  - Struct `SegmentedControl`:
    - `Options []Option` where `type Option struct { Value, Label string; Icon *svg.Icon }`
    - `Selected string`
    - `OnChange func(value string)`
  - Renders `<div class="segmentedcontrol" role="tablist">` with sunken background, active pill, and keyboard arrow navigation.
- `css.go`: Sunken surface track, raised background pill with `0 1px 2px #0000001f` shadow.
- `segmentedcontrol_test.go` and `segmentedcontrol_ui_wasm_test.go`.
- `README.md`.

### Stage 3: `cellgrid` component
Create package `webtyp/components/cellgrid`:
- `cellgrid.go`:
  - Coordinate system: `ColName(col int) string` implementing bijective base-26 (0 -> "A", 25 -> "Z", 26 -> "AA").
  - Struct `CellGrid`:
    - `Cols int`, `Rows int`
    - `SparseCells []CellCoord` (if populated, only renders cells present in this list placed at `grid-row: r+2; grid-column: c+2`)
    - `IsActive func(r, c int) bool`
    - `OnCellClick func(r, c int)`
    - `OnPaintStroke func(cells []CellCoord, on bool)` (receives Bresenham interpolated cells on pointerdrag)
  - Pointer interaction:
    - Handles `pointerdown`, `setPointerCapture`, `elementFromPoint(x, y)`, Bresenham line algorithm (`line(a, b)`), and `pointerup`.
    - Toggle behavior: if drag starts on an active cell, it erases (`on = false`); if on an inactive cell, it paints (`on = true`).
- `css.go`: Fluid grid with `aspect-ratio: 1/1` cells, column and row headers.
- `cellgrid_test.go` and `cellgrid_ui_wasm_test.go`.
- `README.md`.

### Stage 4: Conformance & Catalog
- Register all 3 components in `conformance_test.go`.
- Add catalog entries in `docs/CATALOG.md`.
- Run `gotest ./...`: all tests green.

---

## Acceptance Criteria

1. `gotest ./...` in `webtyp/components` passes with zero failures.
2. `stepindicator`, `segmentedcontrol`, and `cellgrid` pass `conformance_test.go`.
3. No stdlib imports in WASM code.

## Executor notes
The compilation of the three new widgets raises type and unresolved symbol errors related to the `webtyp.com/css` and `webtyp.com/dom` dependencies (e.g. `style.CrossCenter`, `dom.NewEvent`, invalid css.Set tokens). Per explicit user directive, these errors are ignored and the partially broken packages have been committed as-is to allow downstream manual fixes.
