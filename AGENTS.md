# Agent Guide — `webtyp/components`

Constraints for agents adding or modifying UI components. Read this before any change.

---

## Construction Harness — typed & explicit (the WebTyp approach)

This library is part of WebTyp's **construction harness**: the typed, explicit API is what keeps an
agent that doesn't know the library from building wrong code. The compiler must reject mistakes; what
it can't catch becomes a `devMode` warning — never a silent failure.

- **Typed over `any`** — no generic slots; typed builder methods (like `webtyp/json`), reusing `fmt` types. Anything reactive goes only through a signal binding (`BindText`/`Bind*`), which requires a signal.
- **Explicit names** — `Text` (static) vs `BindText` (reactive); reading the call states intent.
- **Illegal states unrepresentable** — dynamic content has ONE path, typed to require a signal.
- **Minimal public surface** — export only what the author types; engine plumbing stays unexported.
- **Docs are minimal "how" instructions, not long skills** — if a rule must be *remembered*, close it
  with types, not prose.

(Ecosystem rationale: skill **api-design** — the construction harness, in
`webtyp/devskills/skills/api-design/SKILL.md`.)

---

## Component Contract — ONE way (signals)

A component implements **only**:

- `Render() *dom.Element` — describes the structure ONCE; dynamic parts are signal bindings.
- `Init(ctx dom.Ctx)` — optional; runs ONCE before first render (load storage, fetch, subscribe).

There is **NO** `OnMount`/`OnUpdate`/`OnUnmount` and **NO** manual `Update()` (it is unexported in
`dom`). Forgetting a lifecycle call is impossible because there are no lifecycle calls.

State the UI shows lives in **typed signals**; changing one patches only the bound DOM node — never
re-render a whole element, never use a Virtual DOM.

```go
type ThemeToggle struct {
    dom.Element                 // value-embed, never *dom.Element
    theme *dom.SignalString     // UI state is a signal, not a plain field
}
func (t *ThemeToggle) Init(_ dom.Ctx) { t.theme = dom.NewString("") /* load saved */ }
func (t *ThemeToggle) Render() *dom.Element {
    return dom.Button().
        BindTextFunc(func() string { return iconFor(t.theme.Get()) }).   // auto-tracked; no deps list
        On("click", func(dom.Event) { t.theme.Set(next(t.theme.Get())) })
}
```

## Component Naming — two words, and the second word must name the class

A component's Go struct name **and** its folder/package name must be at least two words, and the
combination must identify *which style/class* of the thing it is — never just the generic noun alone.

```
✅ ModalDialog (modaldialog/)   ThemeToggle (themetoggle/)   ActionButton (actionbutton/)
❌ Dialog (dialog/)              Toggle (toggle/)              Button (button/)
```

**Why:** a bare generic name (`Dialog`, `Toggle`, `Button`) claims the whole concept for one specific
implementation. If a consumer later needs a different style — a drawer instead of a centered modal, a
segmented switch instead of a click-to-cycle toggle — there's no name left for it without a breaking
rename. Naming the class up front (`ModalDialog`, `DrawerDialog`, `ThemeToggle`, `ThemeSegmented`)
keeps every style addressable and coexisting side by side.

This was violated once: the modal component originally shipped as package `dialog`, struct
`DialogWidget` — a single-word package name for what is really one *style* of dialog. Renamed to
`modaldialog`/`ModalDialog`. Do not repeat this: when creating a component, ask "what specific
style/variant is this?" and put that in the name, not just the generic UI concept.

## No Generics

Zero generic functions in this ecosystem (follow the `webtyp/fmt` codec rule "cero any, cero map").
The DOM boundary is `string`/`bool`, so use concrete typed signals — never `Signal[T]`:

- `SignalString`/`SignalBool`/`SignalNodes`; `NewString`/`NewBool`/`NewNodes`; `Get`/`Set`; `Toggle()` on bool
- Bindings (raw signal): `BindText`, `BindAttr`, `BindClass`, `BindAttrBool`, `Bind` (two-way input),
  `BindChildren` (keyed list — build `[]*Element` in a normal loop, set with `.Key(id)`), `Autofocus`
- Bindings (computed, **auto-tracked, no deps list**): `BindTextFunc`/`BindAttrFunc`/`BindClassFunc`/`BindAttrBoolFunc`;
  `DeriveString`/`DeriveBool` for a named shared computed value
- `Show(boolSig, render)` for conditional subtrees.

## Minimal Public Surface

Export only what a component user types. **Unexport any symbol only this package uses** — e.g. theme
constants (`TsThemeDark`/`TsThemeLight`), helpers like `toggle`/`icon`/`label`. Struct fields stay
unexported; expose behavior through methods.

## WASM / TinyGo — build tags belong to the consumer, not the library

Every `.go` file YOU write is in exactly one of three states. **Every byte of
an untagged file ships to the browser** — the WASM binary must stay minimal.
Some ecosystem libraries (e.g. `webtyp/svg`) never use `//go:build`
internally and instead split backend-only code into a separate importable
sub-package (`webtyp/svg/sprite`) — YOU still choose whether to import that
sub-package from a tagged or untagged file, and that choice is yours to get
right; the library cannot enforce it for you.

| Tag | Compiles into | What belongs there |
|---|---|---|
| *(none)* | WASM **and** backend | `Render()`, signals, typed name constants (icon names, CSS classes) |
| `//go:build wasm` | WASM only | browser-only interaction (`web/client.go`, JS bridges) |
| `//go:build !wasm` | backend/SSR only | CSS embeds (`css.go`), `IconSvg()` + SVG geometry (`svg.go`, imports `webtyp/svg/sprite`), `RenderJS`, heavy HTML strings |

- Lifecycle/reactive code goes in `//go:build wasm` files; provide `!wasm` stubs for any function
  called from tag-less code (e.g. inside `Render()`), returning no-ops / `""`.
- **No Go stdlib** (`fmt`/`strings`/`errors`): use `webtyp.com/fmt`. DOM only via
  `webtyp.com/dom`, never `syscall/js`.
- `switch`, not `map`. No `defer/recover` (a no-op in TinyGo WASM) — use O(1) guards.
- `webtyp/ssr` builds its extractor with the default backend toolchain, so `!wasm` files ARE
  included in SSR extraction — tagging never breaks SSR.

## SVG icons — name is shared, drawing is backend-only

The icon's *name* is the only thing the WASM binary may carry; the geometry is
extracted by `webtyp/ssr` and injected inline into `<body>` by `sitec`
(no `/assets/icons.svg` URL — `href="#id"` resolves without a network request).

- Declare the reference in the untagged component file:
  `const iconX = svg.Icon("comp-x")` (prefix ids with the component name).
  Import `webtyp.com/svg` — safe from any untagged file.
- Define the geometry in `svg.go` under `//go:build !wasm`:
  `sprite.Define(iconX, "0 0 16 16", sprite.Path("..."))` (package
  `webtyp.com/svg/sprite`), returned by `IconSvg() *sprite.Sprite`.
  Always `fill="currentColor"` (Path hardcodes it); color/size are controlled
  from CSS at the use-site.
- Render with `iconX.Render(string(ClsCompIcon))` — never a raw `"#id"` string,
  never hand-built `<svg><use>` (the `svg.Svg()`/`svg.Use()` builders were removed).
- **`webtyp.com/svg/sprite` compiles fine for wasm too** (it's pure
  Go, no build tag of its own) — forgetting `//go:build !wasm` on your
  `svg.go` does NOT fail the build, it silently ships sprite geometry plus
  the `webtyp/json`/`webtyp/model` serialization code to the browser.
  This is caught only by the mandatory pre-publish check, never skip it:

  ```bash
  GOOS=js GOARCH=wasm go list -deps ./... | grep webtyp/svg/sprite   # must be empty
  ```

### A glyph two components share comes from `webtyp/icons`, not a local copy

If the *same* glyph appears in more than one component (the trash/pencil marks
`targetlist`/`targetdate` draw are also `crudview`'s footer buttons), do NOT
define its geometry in one of them and import it sideways, and do NOT paste the
path into each. Import the per-glyph package — `webtyp.com/icons/trash`,
`.../pencil`, … — and take `trash.Ref` (markup) + `trash.Def()` (your
`IconSvg()` sprite). One id, one geometry, assembled everywhere; `assetmin`
collapses the duplicate `Def()`s to one `<symbol>`. A glyph private to one
component still lives in that component's own `svg.go`.

### The *skin* of a shared mark is assembled once too, never re-declared

`listselect` owns BOTH halves of the selection chrome — the markup and the
skin. `RowOf`/`Header` build the row check box and the header strip (its
`partCheck`/`partAll` classes), and `ApplyRow`/`ApplyHeader` paint them; the
`target*` widgets assemble the pieces and never write their own version of
either. Same rule as the geometry: two lists that `crudview` swaps for each
other cannot each own a private copy of the chrome, or they drift. A new
shared mark gets a new `*.Apply`-style helper in the piece that owns it — not
a block pasted into both consumers.

## Testing

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # external agents have no global gotest
gotest
```

- `gotest`, never `go test`. Stdlib assertions only (`testing`/`reflect`, no testify).
- Dual WASM/stdlib via build tags sharing one runner; WASM tests run against a real DOM.
- Cover the **frequent use cases**: load-on-init (no flash), two-way input (IME/cursor safe), derived
  value, conditional `Show`, keyed list, and the no-recursion regression.
- Publish with `gopush 'message'`, never raw `git commit`/`push`.

## Documentation First

Update docs **before** code and before `gopush`: `docs/ARCHITECTURE.md` (what/why),
`docs/DESIGN.md` (decisions), `docs/CATALOG.md`, and the component standards in the `components`
skill (`devflow/skills/components/SKILL.md` — run the `llmskill` sync after editing). `README.md`
must index every file in `docs/`. Diagrams: `flowchart TD`, no `subgraph`, `<br/>` for line breaks.
