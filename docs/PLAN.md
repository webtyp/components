---
PLAN: "feat: components type their fixed UI text as lang.Text and translate it in Render"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 6543375632725665162
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — components: fixed UI text is `lang.Text`

Phase **T4** of the master plan `SOURCE_SELECTION_MASTER_PLAN.md` (orchestration only — everything
this plan needs is inline). **Depends on** `webtyp.com/lang` with `lang.Text` and the page dictionary
(`https://github.com/webtyp/lang/blob/main/docs/PLAN.md`, stages 3b and 4). Do not start before
that tag exists.

Read [AGENTS.md](../AGENTS.md) first (component contract, typed harness). Rules repeated:
- This module compiles to WASM: use `webtyp.com/fmt`, never `strings`/`strconv`/stdlib `fmt`.
- New and rewritten tests live in `tests/` at the module root (external package, public API only).
  A root-level test needs a top-of-file `// Root-level test (justified): …` comment. **Never export a
  symbol so a test can reach it.**
- `lang.Translate` is imported from `webtyp.com/lang`, never `webtyp.com/fmt/lang` (removed).

## Why

Translations are now data: the page carries a dictionary, and `lang.Translate` looks each key up
exactly as written (see the lang plan). A component receives some texts from its caller. Some are
**fixed UI text** (a placeholder, an empty-state message, a dialog title, tab labels), which must be
shown translated. Others are **data** (a patient's name, a conversation title), which must never be
translated: a patient called "Delete" must not become "Eliminar". Today both are `string`, so neither
the component nor the `langc` generator can tell them apart, and modules hard-code Spanish in the
fixed ones.

The type says which is which. Fixed UI text is `lang.Text`. Callers keep writing a plain literal
(`Placeholder: "Search patients"`), because an untyped string constant converts implicitly. The
component translates it in `Render`, and `langc` collects every literal assigned to a `lang.Text`
field.

## Design gate

1. **Prior art.** Flutter separates `String` data from localized getters
   (`AppLocalizations.of(ctx).x`); Angular marks translatable template text with `i18n` attributes,
   and data bindings are never translated; Android uses `@StringRes int` (a resource id) versus a
   plain `String`, so the type tells translatable from data. We follow Android's idea with a string
   type: `lang.Text` versus `string`.
2. **Novice-name test.** `Placeholder lang.Text` — "the placeholder is translatable text". Nothing new
   to learn at the call site.
3. **Complexity ledger.**
   ```
   Concepts the developer must learn   +1 (lang.Text vs string)
   Files they must touch to do X       −1 (no per-app override of component chrome)
   Lines at the call site              +0 for literals; +1 conversion where a caller passes a typed string variable
   Ways to do the same thing           +0
   ```
4. **Where it belongs.** The component knows which of its fields are its own fixed text, so the
   component declares the type and translates at render.
5. **What it deletes.** Translation done by callers (`Placeholder: lang.Translate(x).String()` patterns,
   if any) and the `RegisterWords` uses in this repo's tests.

## Stage 1 — retype the fixed-text fields

Change these fields from `string` to `lang.Text` and add `// English; translated in Render.` to their
comment. In every place the component renders the field (as text or as an attribute such as
`placeholder=`, `aria-label=`, `title=`), use `lang.Translate(c.Field).String()`. An empty field stays
empty and renders nothing, exactly as today.

| File | Type.Field |
|---|---|
| `searchbar/searchbar.go:44` | `SearchBar.Placeholder` (its default `"Search…"` already goes through `Translate`; keep the default as a `lang.Text`) |
| `composebar/composebar.go:27-28` | `ComposeBar.Placeholder`, `ComposeBar.SendLabel` (fix the comment example `"Enviar"` → `"Send"`) |
| `selectsearch/selectsearch.go:102` | `SelectSearch.Placeholder` |
| `bubblethread/bubblethread.go:47` | `BubbleThread.Empty` |
| `inboxlist/inboxlist.go:47` | `InboxList.Empty` |
| `presencelist/presencelist.go:44` | `PresenceList.Empty` |
| `modaldialog/modaldialog.go:31` | `ModalDialog.Title` |
| `decktabs/decktabs.go:40, :70` | `DeckTabs.Label`, `Item.Label` |
| `stepindicator/stepindicator.go:30` | `Step.Label` |
| `segmentedcontrol/segmentedcontrol.go:25` | `Option.Label` |
| `actionbutton/button.go:36` | `ActionButton.Text` |

**Do NOT change** these: they hold data, so they stay `string` and are never translated.
`selectsearch` option `Label` (:95), `inboxlist.Row.Title` (:37), `presencelist.Person.Label` (:35),
`infobar.InfoItem.Text`, `statgrid.StatItem.Label`, `herobanner` `Title`, and `sitenav.NavItem.Label`.
The last three are site content.

Inside this repo, any code that assigns a `string` **variable** to one of the retyped fields now
needs `lang.Text(v)`. Fix those call sites; `go build ./...` lists them.

## Stage 1b — `lang.json` (this library's translations)

Translations are data shipped by each library at its module root. sitec merges them, and the
project's `config/lang.json` wins. Create `lang.json` at the module root with **exactly** this
content. It is library format (no `default`); each value is a positional list in the order of
`languages`, one key per line, and keys are sorted by byte order, as the `langc` generator writes
them. These are the Spanish texts mjosefa-cms registered in Go for searchbar, scheduleeditor and
calendarslider:

```json
{
  "languages": ["es"],
  "keys": {
    "Add": ["Agregar"],
    "Add row": ["Agregar fila"],
    "Add time range": ["Agregar horario"],
    "Back to normal hours": ["Volver al horario normal"],
    "Blocked": ["Bloqueado"],
    "Calendar": ["Calendario"],
    "Date": ["Fecha"],
    "Dates that differ from the weekly pattern": ["Fechas que se apartan del patrón semanal"],
    "Does not work": ["No atiende"],
    "Extra day": ["Día extra"],
    "Fri": ["Vie"],
    "From": ["Desde"],
    "Hours for marked days": ["Horario de días marcados"],
    "I do not work that day": ["Ese día no atiendo"],
    "Marked days": ["Días marcados"],
    "Mon": ["Lun"],
    "Next month": ["Mes siguiente"],
    "No exceptions": ["Sin excepciones"],
    "No specific dates yet": ["Todavía no hay fechas especiales"],
    "Notes": ["Notas"],
    "Previous month": ["Mes anterior"],
    "Reason": ["Motivo"],
    "Remove": ["Quitar"],
    "Remove row": ["Quitar fila"],
    "Remove time range": ["Quitar horario"],
    "Sat": ["Sáb"],
    "Save these hours": ["Guardar ese horario"],
    "Search…": ["Buscar…"],
    "Select date": ["Elegir fecha"],
    "Special hours": ["Horario especial"],
    "Specific dates": ["Fechas específicas"],
    "Sun": ["Dom"],
    "Thu": ["Jue"],
    "To": ["Hasta"],
    "Today": ["Hoy"],
    "Tue": ["Mar"],
    "Turn on the days you work and set their hours": ["Marque los días que atiende y fije su horario"],
    "Type": ["Tipo"],
    "Wed": ["Mié"],
    "Weekly pattern": ["Patrón semanal"],
    "Work that day": ["Atender ese día"],
    "You do not work this weekday": ["Ese día de la semana no atiende"],
    "occupied": ["ocupado"],
    "or work different hours that day": ["…o atender en otro horario"]
  }
}
```

## Stage 2 — tests

- New file `tests/text_fields_test.go` (backend, `package components_test` or per-package external
  test packages under `tests/`, matching the existing `tests/` layout if any). For each retyped
  component, render it with a literal value and assert the English text appears as is (no dictionary
  on the backend: pass-through).
- New file `tests/text_fields_translate_test.go` (`//go:build wasm`). `TestMain` inserts
  `<script type="application/json" id="` + lang.ScriptID + `">` with
  `{"default":"es","languages":["es"],"keys":{"Search patients":["Buscar pacientes"],"Send":["Enviar"],"Hours":["Horario"]}}`
  before any lookup. Under `lang.OutLang(lang.ES)`, a `SearchBar{Placeholder: "Search patients"}`
  renders `placeholder="Buscar pacientes"`, `ComposeBar{SendLabel: "Send"}` shows `Enviar`,
  and a `DeckTabs` item labelled `Hours` shows `Horario`. A data field (a `selectsearch` option
  `Label: "Delete"`) is shown as `Delete` even with `"Delete"` in the dictionary.
- `tests/lang_json_test.go` (backend): `lang.json` parses as JSON, and every list has exactly
  `len(languages)` values.
- `searchbar/searchbar_test.go` and `selectsearch/selectsearch_test.go` call `lang.RegisterWords` /
  `lang.OutLang(ES)`, and `RegisterWords` no longer exists. Move their translation cases into the
  WASM test above, and delete those cases from the package test files. Leave the rest of those files
  as they are.

## Acceptance

- `gotest` passes (includes WASM).
- `grep -rn "RegisterWords\|DictEntry\|webtyp.com/fmt/lang" --include='*.go' .` → empty.
- `grep -rnE "^\s+(Placeholder|SendLabel|Empty) +string" --include='*.go' searchbar composebar selectsearch bubblethread inboxlist presencelist` → empty.
- `## Executor notes` lists every call site that needed a `lang.Text(v)` conversion.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Retype + translate | the files in the table, plus call sites `go build` reports, `go.mod`, `go.sum` |
| 1b | `lang.json` | `lang.json` |
| 2 | Tests | `tests/lang_json_test.go`, `tests/text_fields_test.go`, `tests/text_fields_translate_test.go`, `searchbar/searchbar_test.go`, `selectsearch/selectsearch_test.go` |
