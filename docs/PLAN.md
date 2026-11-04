---
PLAN: "feat: agedate — a birth date field that also accepts the age"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 1332858877940491367
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — `agedate.AgeDate`: type the age, get the birth date

> The files `docs/PLAN_STAGE_*` belong to another plan: ignore them, do not edit or delete them.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.

## Why

At a clinic's front desk people are asked their **age**, not their birth date. The previous system had a field
that did both: typing `40` filled the birth date 40 years back, typing a date showed the age; only the **date** was
stored, because an age changes every year. This plan rebuilds it in Go for any webtyp form, without JavaScript.

## Ecosystem search

- `form.Renderer` (`webtyp.com/form`, `render_input.go:82`): `RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element` — an input that draws its own control; reference implementer `form/tests/renderer_test.go` (`customInput` embedding `input.Base`) → **reuse**.
- `input.Date()` (`webtyp.com/input/date.go`): YYYY-MM-DD validation → **reuse** for `Validate`.
- No age field exists in `input` or `components` (grep `age`, `birth`) → **new**.

## Design gate

1. Prior art: HL7 FHIR stores `Patient.birthDate`, never the age (https://hl7.org/fhir/patient.html); the age is
   computed. Followed: the value is only the date.
2. Name: `AgeDate` (`<characteristic><generic>`: a date typed by age; sibling would be a plain date picker).
3. Ledger: +1 component, +1 constructor `agedate.New()`. Ways to enter a date: `input.Date()` (plain) or
   `agedate.New()` (with age); same stored format.
4. Belongs in `components`: it renders DOM (`input` stays free of `dom`).
5. Deletes nothing.

## Stage 1 — `agedate/agedate.go`

```go
package agedate

const NameAgeDate = widget.Name("agedate")
const (
	PartAge  = widget.Part("age")
	PartDate = widget.Part("date")
)

// AgeDate is a birth date field with an age box beside it: typing the age fills the date (today minus that age),
// choosing the date shows the age. Only the date (YYYY-MM-DD) is the value. An age of 0.5 means 6 months: the
// decimal part is a fraction of a year (months = floor(fraction × 12)).
//
// Example (in a model.Definition): {Name: "birthdate", Type: agedate.New(), Label: "Birthdate", OmitEmpty: true}
type AgeDate struct {
	input.Base
	date input.Input // input.Date(): validation of the stored value
}

func New() *AgeDate // InitBase("", "", "date"); date = input.Date()

func (a *AgeDate) Clone(parentID, name string) input.Input
func (a *AgeDate) Validate(value string) error // "" allowed when not required; otherwise a.date.Validate(value)
func (a *AgeDate) RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element
func (a *AgeDate) WidgetName() widget.Name { return NameAgeDate }
func (a *AgeDate) WidgetKind() widget.Kind // the Kind of a form control, the same one input fields use in this repo
```

Pure functions (same file, unexported, tested through the exported behaviour below):

```go
// birthFromAge: today minus years and months. ageText uses '.' or ',' as decimal mark. "" or "0" → "", false.
func birthFromAge(ageText string, today time.Unix) (date string, ok bool)
// ageFromBirth: whole years, plus the year fraction of full months, rounded to 1 decimal ("40.3").
func ageFromBirth(date string, today time.Unix) (age string, ok bool)
```

`RenderInput`: a root `div` (`NameAgeDate.Root()`) with two labelled controls: an `input type=number min=0 max=150
step=0.1` (`PartAge`, label `lang.Text("Age")`, help-free) and an `input type=date` (`PartDate`) bound to `value`.
Age `input` event → `birthFromAge(v, time.NowUnix())` → set the date control and call `onInput(date)`. Date `change`
event → `onInput(date)` and set the age box with `ageFromBirth`. Initial render: when `value` holds a date, fill the
age box. Use `webtyp.com/time` (never stdlib `time`), no `map`, ids minted by `dom` (skill dom-elements).

## Stage 2 — `agedate/css.go` (`//go:build !wasm`)

Two controls on one row (age narrow, date wide) using the existing `widget/style` recipes for form fields, as other
components' `css.go` do. A recipe that is missing is a defect to report in the PR, never hand-written CSS.

## Stage 3 — register in `conformance_test.go` (all three places, per the components skill)

## Stage 4 — tests (`agedate/agedate_test.go`, `//go:build !wasm`) — behaviour only

| Test | Call (today = 2026-11-03) | Expected |
|---|---|---|
| `TestBirthFromAgeYears` | `"40"` | `"1986-11-03"` |
| `TestBirthFromAgeMonths` | `"0.5"` and `"0,5"` | `"2026-05-03"` |
| `TestBirthFromAgeEmpty` | `""`, `"0"` | `ok == false` |
| `TestAgeFromBirth` | `"1986-11-03"` | `"40"` or `"40.0"` (document which) ; `"2026-05-03"` → `"0.5"` |
| `TestValidate` | `"2026-02-30"` | error ; `"1986-11-03"` → nil |

Expose the two pure functions to the test **without exporting them**: the test file is in package `agedate`
(internal test), allowed here because they are pure date arithmetic.

## Stage 5 — docs and translations

`langc sync .`, Spanish values ("Age" → "Edad"). `agedate/README.md`: why (front desk asks the age), the example,
and "only the date is stored". Add the component to the repo README's list.

## Delivery check

1. `gotest` passes (a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `grep -rn '"time"' agedate/` → empty (only `webtyp.com/time`).
3. `langc sync .` → 0 untranslated `es`.

Commit after every stage. Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own
work. Stuck 30 minutes: open the PR anyway listing what fails.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `agedate/agedate.go` | compiles |
| 2 | `agedate/css.go` | compiles |
| 3 | `conformance_test.go` | conformance passes |
| 4 | `agedate/agedate_test.go` | tests pass |
| 5 | `agedate/README.md`, `README.md`, `lang.json` | done |
