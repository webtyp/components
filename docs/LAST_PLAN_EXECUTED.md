---
PLAN: "fix: replace inline style attributes with dom.StyleProp in calendarslider and cellgrid"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan: `components` sin atributos HTML `style=""` (CSP `style-src 'self'`)

Master: `SECURITY_OBSERVABILITY_MASTER_PLAN.md` (D3, O3b-5).
**Prerrequisito:** `webtyp.com/dom` con `Element.StyleProp` (O3b-1).

## Tools

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
go install webtyp.com/devflow/cmd/gotest@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Ecosystem search

- `grep -rn '"style"' components/ --include='*.go'` muestra exactamente 3 sitios:
  - `calendarslider/calendarslider.go:610`
  - `cellgrid/cellgrid.go:57`
  - `cellgrid/cellgrid.go:83`
- **Decisión:** migrar los 3 a `dom.Element.StyleProp` (CSSOM) y mover `display: grid` de `cellgrid` a su hoja CSS.

## Design gate

No cambia ninguna firma pública de `webtyp.com/components`.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `(*dom.Element).StyleProp` | `func (b *Element) StyleProp(property, value string) *Element` | *(O3b-1)* | `webtyp.com/dom` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Reglas de entrega y anti-footguns

- `calendarslider` y `cellgrid` compilan a WASM con TinyGo: usar `webtyp.com/fmt`, nunca stdlib `fmt`/`strings` ni `map`.
- Hacer commit después de cada etapa. No usar `git reset`, `git restore`, `git checkout -- <file>` ni `git stash`.

---

## Etapa 1: Pruebas (TDD)

Agregar prueba en `tests/no_inline_style_test.go` (y registrar en `docs/TESTS.md` bajo `| Test | Rule |`):
- `TestComponents_NoInlineStyleAttr`: renderizar `CalendarSlider` (con días seleccionables) y `CellGrid{Rows: 2, Cols: 3}` y verificar que el HTML resultante (`String()`) NO contiene `style=`.

## Etapa 2: Migrar `calendarslider` y `cellgrid`

1. En `calendarslider/calendarslider.go:608-611`:
   Reemplazar:
   ```go
   Attr("style", "--meter-fill:"+fmt.Sprint(use)+"%;")
   ```
   por:
   ```go
   StyleProp("--meter-fill", fmt.Sprint(use)+"%")
   ```
2. En `cellgrid/cellgrid.go:57`:
   Reemplazar:
   ```go
   grid.Attr("style", fmt.Sprintf("display: grid; grid-template-columns: repeat(%d, 1fr); grid-template-rows: repeat(%d, 1fr);", c.Cols, c.Rows))
   ```
   por:
   ```go
   grid.StyleProp("grid-template-columns", fmt.Sprintf("repeat(%d, 1fr)", c.Cols)).
       StyleProp("grid-template-rows", fmt.Sprintf("repeat(%d, 1fr)", c.Rows))
   ```
   Y asegurar que `cellgrid/css.go` declara `display: grid` en `Root(...)` (vía receta de `widget/style` o regla de layout grid).
3. En `cellgrid/cellgrid.go:83`:
   Reemplazar:
   ```go
   Attr("style", fmt.Sprintf("grid-row: %d; grid-column: %d;", r+1, col+1))
   ```
   por:
   ```go
   StyleProp("grid-row", fmt.Sprint(r+1)).
       StyleProp("grid-column", fmt.Sprint(col+1))
   ```

## Etapa 3: Verificación y entrega

1. Verificar con grep: `grep -rn '"style"' . --include='*.go' | grep -v '_test.go'` → vacío.
2. Ejecutar `gotest`.

## Tabla de etapas

| Etapa | Archivos | Qué |
|---|---|---|
| 1 | `tests/no_inline_style_test.go`, `docs/TESTS.md` | Prueba TDD sin `style=` |
| 2 | `calendarslider/calendarslider.go`, `cellgrid/cellgrid.go`, `cellgrid/css.go` | Reemplazo por `StyleProp` y CSS |
| 3 | `docs/TESTS.md` | Verificación grep y `gotest` verde |
