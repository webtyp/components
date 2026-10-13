---
PLAN: "feat(themetoggle): 3-state segmented control + usermenu Preferences slot"
STATUS: running
SESSION: 2116987196617954533
---

# PLAN — themetoggle: 3 states + usermenu Preferences slot

Read `AGENTS.md` first (WASM/TinyGo rules: no maps, `webtyp.com/fmt` not stdlib fmt, style DSL only, tests under `tests/`).

## Goal
`webtyp/layout/platformd` will embed the theme control inside the user menu panel of every app (done later, outside this repo). This plan prepares the two components it needs.

## Stage 1 — themetoggle (`themetoggle/`)
1. States: `TsThemeSystem = "system"`, `TsThemeLight = "light"`, `TsThemeDark = "dark"`. `valid()` accepts all three.
2. Render a segmented control instead of one button: root `Div` with `role="radiogroup"`, `aria-label="Theme"`; three `Button`s (part `PartOption = widget.Part("option")`), order system, light, dark. Each has `title` + `aria-label` (`"System"`, `"Light"`, `"Dark"` — lang keys via `webtyp.com/lang` as in the rest of the repo), `role="radio"`, and `aria-checked` bound to the signal (`BindAttr` with a `DeriveString` returning `"true"`/`"false"`). Use `BindStateFunc(widget.Current, ...)` on the selected option so CSS can style it. Keep the `supported == false → Span()` branch.
3. Icons: no emoji. Add `themetoggle/svg.go` (`//go:build !wasm`) with `func (t *ThemeToggle) IconSvg() *sprite.Sprite` defining three outline glyphs (viewBox `0 0 24 24`, stroke-style paths from Lucide, MIT): `IconSystem = svg.Icon("themetoggle-system")` (monitor), `IconLight = svg.Icon("themetoggle-light")` (sun), `IconDark = svg.Icon("themetoggle-dark")` (moon). Declare the `svg.Icon` consts in `themeswitch.go` (WASM-safe). Render with `Icon.Render(class)`. Follow the pattern of any other component in this repo that has an `svg.go`.
4. Behaviour (`themeswitch_wasm.go`):
   - Init: stored valid value → apply with `SetDocumentAttr("data-theme", v)` and select it. No stored value → do NOT write `data-theme`; select `light`/`dark` from `DarkSchemeActive()` (this reflects an app's `css.DefaultLight()` correctly).
   - Click on an option: `SetDocumentAttr("data-theme", v)` (also for `"system"`), `LocalStorageSet(storageKey, v)`, set signal. Same `storageKey`.
   - Backend (`themeswitch_backend.go`): default signal `light`, `supported = true`, `onClick` no-op (signature now takes the chosen `TsTheme`).
5. CSS (`css.go`): remove `Docked(Viewport, ...)`, `Button(Primary)` and `Round(RadiusFull)` — position belongs to the host. Root: `style.Row(style.Space1)`, `style.As(style.Inset)`, `style.Round(style.RadiusFull)`, small padding. Option part: `style.Button(style.Subtle)`, `style.Round(style.RadiusFull)`, `style.Pad(style.Space1)`, icon box `style.IconBox(style.IconSm)`; current state: `style.As(style.Panel)` (raised surface). Icons inherit `currentColor`; unselected options muted. If a needed recipe is missing in `webtyp/widget/style`, STOP and report it in the PR description — do not hand-write CSS.
6. Update `themetoggle/README.md` (3 states, not floating, sits wherever the host puts it) and the component demo `themetoggle/web/client.go` if it relies on the floating position.

## Stage 2 — usermenu Preferences slot (`usermenu/`)
1. New field `Preferences Component` on `UserMenu` (doc: user preferences the shell owns, e.g. theme; `Actions` stays for app actions like sign-out).
2. New part `PartPrefs = widget.Part("prefs")`; render `Div().Set(clsPrefs)` with the component BETWEEN roles and actions, only when non-nil.
3. CSS: `Part(PartPrefs, style.Row(style.Space2), style.CenterContent())` (or the closest existing recipes).

## Tests (write first, see them fail, then implement) — under `tests/`
- themetoggle: `valid` accepts system/light/dark and rejects others; rendered markup has 3 options with `role="radio"`, exactly one `aria-checked="true"`; no emoji characters in the markup; stylesheet contains no `position: fixed`.
- usermenu: Preferences renders inside the panel, before the actions block; omitted when nil.
- Existing tests (`themeswitch_test.go`, `uc_themeswitch_test.go`, contract test, usermenu anchor test) must be updated, not deleted, and pass.
- Run `gotest` (stdlib + wasm) at repo root; all green.

## Out of scope
`webtyp/css` reset rule for `[data-theme="system"]` and the platformd wiring — done separately.
