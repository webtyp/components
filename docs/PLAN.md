---
PLAN: "feat(components)!: section titles use style.Heading; typed ActionButton variant; remove decktabs"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — components follow the one title and one way to group screens

**Requires** `webtyp.com/widget` with `style.Heading()` (plan "feat(style): Heading recipe"). Run
`go get webtyp.com/widget@v0.6.41` first; if `style.Heading` does not exist, stop.

## Problem

Section titles are hand-sized per component (`stepaccordion` title `TextBase`+`WeightMedium`,
`contentcard` header `WeightBold`, `modaldialog` an unstyled `H2`), `ActionButton.Variant` is a free
string where a typo silently renders the primary style, and `decktabs` is a second way to group screens
that the shell already provides with `platformd.NewGroupUIModule` (decided D6/P16 in master plan
`UI_CONSISTENCY`).

## Design gate

1. **Prior art.** Material 3 / Polaris type roles (components ask for a role, not a size); Polaris
   `Button variant` and Material `ButtonStyle` are closed enums, not strings; Android/iOS group screens
   in the navigation shell (bottom nav / tab bar), not in an in-page tab strip.
2. **Novice-name.** `actionbutton.Variant` with constants `Primary`, `Secondary`, `Danger` — the words
   the field's comment already uses.
3. **Ledger.** Concepts −1 (`decktabs`). Ways to group screens 2 → 1. Ways to size a section title N → 1.
   Invalid variant strings: representable → not compilable.
4. **Where it belongs.** Title size: `widget/style.Heading`. Variant: `actionbutton`. Grouping: `platformd`.
5. **What it deletes.** Package `decktabs` (all files), its entries in `conformance_test.go` and in
   `tests/text_fields_test.go`, `tests/text_fields_translate_test.go`; the hand-composed
   `FontSize`/`FontWeight` on the title parts below; the `string` variant switch.

## Stage 1 — titles

| File | Part | Replace with |
|---|---|---|
| `stepaccordion/css.go` | `PartTitle` | `style.Heading()` |
| `contentcard/css.go` | `PartHeader` | keep `Pad(Space2)`, replace `FontWeight(WeightBold)` with `style.Heading()` |
| `modaldialog/modaldialog.go` + `css.go` | the `H2` title | give it a part `PartTitle = widget.Part("title")` set with `.Set(cls.AsAttr())`, styled `style.Heading()` |

The step title stays an `<h3>` → change it to `<h2>`, and the modal title stays `<h2>`: inside content
there is one heading level (the page `<h1>` is the platform header's).

## Stage 2 — typed variant (`actionbutton/button.go`)

```go
// Variant is the weight of the action: Primary for the main one, Secondary
// for alternatives, Danger for destructive ones.
type Variant uint8

const (
	Primary Variant = iota
	Secondary
	Danger
)
```

- `ActionButton.Variant Variant`; the `switch` matches the constants; zero value = `Primary`.
- Tests and README use `actionbutton.Secondary` etc.; delete the `[]string{"primary", …}` loop in
  favour of `[]actionbutton.Variant{Primary, Secondary, Danger}`.

## Stage 3 — remove `decktabs`

Delete `decktabs/`; remove it from `conformance_test.go` (import, slice, Kind map) and from the two
`tests/text_fields*_test.go` cases. README index: remove the row.

## Stage 4 — docs

- `stepaccordion`, `contentcard`, `modaldialog`, `actionbutton` READMEs: titles use `style.Heading`;
  variant constants.
- No new tests: titles and variants are look (the typed `Variant` is enforced by the compiler). Delete
  the `decktabs` tests and any `actionbutton` test that only asserts class strings per variant.

## Acceptance

- `grep -rn "decktabs" .` → empty; `grep -rn 'Variant: "' --include=*.go .` → empty.
- `gotest ./...` green.

| Stage | Files |
|---|---|
| 1 | `stepaccordion/stepaccordion.go`, `stepaccordion/css.go`, `contentcard/css.go`, `modaldialog/modaldialog.go`, `modaldialog/css.go` |
| 2 | `actionbutton/button.go`, `actionbutton/*_test.go`, `actionbutton/README.md` |
| 3 | `decktabs/` (deleted), `conformance_test.go`, `tests/text_fields_test.go`, `tests/text_fields_translate_test.go`, `README.md` |
| 4 | READMEs above, tests |
