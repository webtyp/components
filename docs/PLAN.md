---
PLAN: "feat!: list widgets read view.Item.Badge; targetdate and targethour draw the date and hour from it"
TAG: v0.8.56
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 6444490990618543278
PR: https://github.com/webtyp/components/pull/34
REVIEWED_COMMIT: 275e8ae5ff3b23c48bd9a7f6e2a7aac4a71d7096
---

> Read AGENTS.md at the repo root before any change.

<!-- Dispatch only after webtyp.com/view v0.6.47 is published (it removes Item.LeadTop/LeadMain/LeadBottom and
     adds Item.Badge). -->

# Plan — `webtyp.com/components`: the row badge comes from one field

## Why

`webtyp.com/view` v0.6.47 builds every list row from the record's fields (`model.RowOf`): `Item{ID, Label,
Description, Badge}`. `Badge` is the raw text of the one field marked `model.RowBadge` — for a date or an hour it
is the field's stored value (seconds since 1970, `webtyp.com/time` `Unix`, as text), or an already formatted text
(«09:30»). The three hand-filled lines `LeadTop/LeadMain/LeadBottom` are gone: each module used to format them by
hand (with untranslated Spanish abbreviations in `clinical_encounter/view.go:30-47`). The widgets that draw a
leading badge now format it themselves, translated, once.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/view@v0.6.47
go get webtyp.com/time@v0.6.2
```

If `view@v0.6.47` still has `Item.LeadTop` or lacks `Item.Badge`, stop and say so in the PR.

## Ecosystem search

- `time.CivilDate(u Unix) (year, month, day, hour, min, sec int)`, `time.Weekday(u Unix) int`,
  `time.FormatTime(u Unix) string` (`webtyp.com/time` v0.6.2, `api.go`) → **reuse**.
- `fmt.Convert(s).Int64()` (`webtyp.com/fmt`) to read the badge text as a number → reuse (if `Int64` does not exist
  with that shape, stop and name it in the PR).
- `lang.Text` + this repo's `lang.json` for the weekday/month abbreviations → reuse; no hand-written Spanish arrays.

## Design gate

1. **Prior art.** Material's list «leading element» and iOS's date badges take one date value and format it in the
   component, in the user's locale; Django admin formats a `DateField` column itself.
2. **Novice-name test.** `Item.Badge` — «the badge of the row». No new exported name here.
3. **Complexity ledger.** Concepts −3 (`LeadTop/LeadMain/LeadBottom`); date formatting written once (here) instead
   of per module.
4. **Where it belongs.** Drawing a date badge is the widget's job.
5. **What it deletes.** Every read of `LeadTop/LeadMain/LeadBottom`; the per-module date formatting (in the module plans).

## Stage 1 — `targetdate`: badge → three lines (`targetdate/targetdate.go`)

Add (keep the Parts `PartLeadTop/Main/Bottom` and their classes: they are the three drawn lines):

```go
// Abbreviations, translated through lang.json (English keys).
var (
	weekdayAbbr = [7]lang.Text{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	monthAbbr   = [12]lang.Text{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
)

// leadLines turns the row badge into the three lines of the date badge: a stored
// date (seconds since 1970, as text) → weekday / day / month; any other text → the
// text alone in the middle line; "" → three empty lines.
func leadLines(badge string) (top, main, bottom string) {
	if badge == "" {
		return "", "", ""
	}
	secs, err := fmt.Convert(badge).Int64()
	if err != nil {
		return "", badge, ""
	}
	u := time.Unix(secs)
	_, month, day, _, _, _ := time.CivilDate(u)
	return lang.Translate(weekdayAbbr[time.Weekday(u)]).String(),
		fmt.Convert(day).String(),
		lang.Translate(monthAbbr[month-1]).String()
}
```

Where the row state is built (lines ~72-84) use `top, main, bottom := leadLines(it.Badge)` instead of
`it.LeadTop/LeadMain/LeadBottom`. Update the doc comments (lines 2, 52, 88) to say «the badge of the row
(`view.Item.Badge`)». Use the `time` and `fmt` import paths of `webtyp.com/*`, never the stdlib.

## Stage 2 — `targethour`: badge → hour (`targethour/targethour.go`)

Same pattern: the main line is `hourText(it.Badge)`:

```go
// hourText: a stored instant (seconds since 1970, as text) → «HH:MM» in the
// configured zone; any other text (already «09:30») → as is.
func hourText(badge string) string {
	secs, err := fmt.Convert(badge).Int64()
	if err != nil {
		return badge
	}
	return time.FormatTime(time.Unix(secs))
}
```

Every former read of `LeadTop`/`LeadMain`/`LeadBottom` in `targethour` uses `hourText(it.Badge)` for the hour line
and `""` for the others. Check: `grep -rn "LeadTop\|LeadMain\|LeadBottom" --include=*.go .` → only the `Part`/class
identifiers named `PartLead*`/`clsLead*` remain.

## Stage 3 — `targetlist` comment

`targetlist/targetlist.go:39`: «itself only ever reads ID/Label/Description — Badge is for the date and hour lists».

## Stage 4 — tests

Update every test that built `view.Item{LeadTop:…}` (`targetdate/row_test.go`, `targethour/*_test.go`) to
`view.Item{Badge: …}`. Add:

| Test | Input `Badge` | Expected |
|---|---|---|
| `TestLeadLines_StoredDate` (targetdate) | text of `time.Date(2026, 10, 27, 12, 0, 0)` | top = translated «Tue», main «27», bottom = translated «Oct» |
| `TestLeadLines_PlainText` | `"Hoy"` | main «Hoy», top/bottom «» |
| `TestLeadLines_Empty` | `""` | three «» |
| `TestHourText_StoredInstant` (targethour) | text of an instant at 09:30 in the configured zone | `time.FormatTime` of it |
| `TestHourText_PlainText` | `"09:30"` | `"09:30"` |

## Stage 5 — docs and translations

- `README.md` of the repo and of `targetdate`/`targethour` if they mention `Lead*`: the row's `Badge`.
- `langc sync .`, fill the 19 new keys (7 weekdays, 12 months) for every language in `"languages"`. Zero
  untranslated keys, zero reuse warnings.

## Delivery

- `gotest` green (also the `-tags browser` tests in headless Chrome). A PR that does not compile, or whose tests
  could not run, is a failed delivery. Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`,
  `GOFLAGS=-mod=mod`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `targetdate/targetdate.go`, `go.mod`, `go.sum` | `leadLines` from `Badge` |
| 2 | `targethour/targethour.go` | `hourText` from `Badge` |
| 3 | `targetlist/targetlist.go` | comment |
| 4 | `targetdate/*_test.go`, `targethour/*_test.go` | badge tests |
| 5 | `README.md`, `lang.json` | document, translate |
