---
PLAN: "fix(calendarslider): instance-prefix month/day ids so two calendars on one page cannot collide"
TAG: v0.6.18
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 5533973471740343706
---

# PLAN — calendarslider: ids globales colisionan entre instancias

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` (Etapa D) — el módulo
`agenda` de `app-demo` monta un `ScheduleEditor` que internamente renderiza un
`calendarslider`, Y el módulo `reservation` ya monta otro `calendarslider`
como filtro del `crudview`. Ambos viven en el mismo render de
`platformd.Render()` (el chasis renderiza TODOS los módulos montados en un solo
árbol — ver `webtyp/layout/platformd`), y ambos emiten **el mismo id** para el
mes `2026-09`:

```
panic: dom: id cs-m-2026-09 was written twice in one render, by <div> and <div> —
a single component instance is being rendered in two places, so one copy is inert
```

Incluso sin dos calendarios, una instancia única ya es frágil: el id no es
instancia-específico, depende del orden de montaje/serialización.

## Causa raíz

`calendarslider/calendarslider.go` construye ids **globales**:

- `buildMonth` → `ID("cs-m-" + key)` (línea ~341)
- botones ‹ › → `Attr("data-target", "cs-m-"+prevKey/nextKey)` (378/385)
- `slideToMonth` → `Get("cs-m-" + key)` (404)
- `buildDay` → `ID("cs-d-"+dateStr)` (488)

Mientras que `c.uid` (line ~183) YA existe y se usa SOLO para el toggle
colapsado (`"cs-"+...` no incluye el uid). El componente nació pensando "un
calendario por página"; la demo (y cualquier app con dos calendarios en el
mismo árbol) lo rompe.

La regla que se viola es la del `doc.go` / `CONSTRUCTION_HARNESS`: **un
componente no debe emitir ids clonables entre instancias**; dos instancias en
el mismo render deben tener ids disjuntos.

## Cambio

Prefijar TODOS los ids de instancia con `c.uid`, de modo que dos
`CalendarSlider` no colisionen jamás:

1. `buildMonth`:
   - `ID(c.uid+"-m-"+key)`
   - `Attr("data-target", c.uid+"-m-"+prevKey)` y `-m-`+nextKey
   - los onclick siguen llamando a un `slideToMonth` pero con uid:
     `prev.On("click", ...)` → `c.slideToMonth(prevKey, prevWraps)`
2. `slideToMonth` pasa a ser **método de instancia**:
   ```go
   func (c *CalendarSlider) slideToMonth(key string, instant bool) {
       ref, ok := Get(c.uid + "-m-" + key)
       ...
   }
   ```
   (es el patrón correcto: el uid pertenece a la instancia; la función libre
   con id global es exactamente el hueco.)
3. `buildDay` → `ID(c.uid+"-d-"+dateStr)` (el `data-date` queda igual: es un
   atributo semántico, no un id de nodo).
4. `buildCollapsed` NO cambia (ya usa `c.uid+suffixCollapsedToggle`).

Compatibilidad: el prefijo cambia los `id`/`data-target`/lookup de forma
consistente en las 3 piezas, así que la navegación sigue funcionando. Los tests
existentes que asumen `cs-m-`/`cs-d-` se actualizan (ver abajo). Nada más del
árbol depende de `"cs-m-"` (grep: solo este fichero).

## Tests

Actualizar `calendarslider_test.go` / `calendarslider_wasm_test.go`:

- Los asserts de id `cs-m-…`/`cs-d-…` pasan a `UID + "-m-" + …` /
  `UID + "-d-" + …` — ver cómo se construye el uid en `Init`
  (estable por instancia gracias a `nextCalendarSliderID`).
- **Nuevo test de regresión**: montar DOS `CalendarSlider` en el mismo render
  (via `dom.Render`/serialización con observador) y verificar que NO haya
  ids duplicados entre ambos (cada mes de cada calendario con su prefijo).
  Este test es el que la demo necesita: si el chasis monta N módulos con
  calendar, N calendarios deben coexistir.
- `gotest ./calendarslider/` verde.

## Criterios de aceptación

- `gotest ./...` verde en `components`; `GOOS=js GOARCH=wasm go build ./...` OK.
- La demo (`app-demo` con `agenda` + `reservation` montados) carga sin panic
  de `claimID`, con reserva y agenda visibles.
- `go list -deps ./calendarslider/ | grep webtyp/svg/sprite` vacío.

## Fuera de alcance

- Cambiar la forma de `data-date`/`data-target` (siguen siendo atributos).
- Hacer el `uid` configurable por el host (no hace falta: el generador interno
  es suficiente y estable dentro de una sesión).