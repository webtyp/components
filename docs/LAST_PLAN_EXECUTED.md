---
PLAN: "feat: scheduleeditor component + targethour free-slot rows (demo agenda feature)"
TAG: v0.7.0
EXECUTOR: local
REVIEWER: none
---

# PLAN — `components` para la feature "Agenda + Reserva" (Etapa A del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §4.3, §7 fila A.
(Copia local: `/home/cesar/Dev/Project/webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md`.)

**Dos partes independientes** en un solo `PLAN.md` (patrón `RESERVATION_VIEW_FIXES`):

- **Parte 1** — componente nuevo `scheduleeditor`.
- **Parte 2** — `targethour` gana filas de "hueco libre" (`FreeSlots`).

No hay dependencia entre P1 y P2; se pueden ejecutar en cualquier orden. Ambas
son necesarias para la demo (P1 para la Etapa D, P2 para la Etapa G).

> Nota: la skill `components` cargada en algunas sesiones está **desactualizada**
> (menciona `OnMount()` y `ssr.go`). La autoridad es `components/AGENTS.md` +
> el código real: contrato = `Render()` + `Init(ctx dom.Ctx)` (sin `OnMount`),
> CSS en `css.go` y SVG en `svg.go` (ambos `//go:build !wasm`), nunca `ssr.go`.
> Referencia viva: `components/targethour/` (`targethour.go` + `css.go` + `svg.go`).

---

# Parte 1 — `components/scheduleeditor`

## Objetivo

Componente **puro** (`Render()` + `Init()`, cero `router`/`orm`/módulos de
dominio) para editar la agenda de un profesional: una **plantilla semanal** de 7
filas y un **panel de excepciones por fecha**. El host traduce los callbacks a
ops de `appointment_booking` (Etapa C/D) — el componente no lo sabe.

Forma de datos = minutos int desde medianoche (forma `appointment_booking`).
Justificación: master plan §8.

## Ficheros (paquete `scheduleeditor/`, plano)

```
components/scheduleeditor/
  scheduleeditor.go        # tipos públicos, ScheduleEditor, Render(), Init()
  css.go                   # //go:build !wasm  — //go:embed scheduleeditor.css, RenderCSS()
  scheduleeditor.css
  svg.go                   # //go:build !wasm  — IconSvg() *sprite.Sprite (patrón targethour/svg.go)
  scheduleeditor_test.go   # backend (sin build tag)
  scheduleeditor_ui_wasm_test.go  # //go:build wasm — interacción
```

## API pública (fijada por el master plan §4.3 — no desviarse)

```go
package scheduleeditor

type WeeklyRow struct {
    DayOfWeek               int  // 0=Domingo … 6=Sábado
    Active                  bool
    WorkStart, WorkFinish   int  // minutos 0..1439
    BreakStart, BreakFinish int  // 0/0 = sin colación
}

type Exception struct {
    ID              string
    Date            string // "YYYY-MM-DD"
    Type            string // ExcHoliday | ExcSpecialHours | ExcBlocked
    StartMin, EndMin int
    Notes           string
}

const (
    ExcHoliday      = "HOLIDAY"
    ExcSpecialHours = "SPECIAL_HOURS"
    ExcBlocked      = "BLOCKED"
)

type ScheduleEditor struct {
    dom.Element
    Week              []WeeklyRow      // el host pasa 7 filas (Dom..Sáb) ya ordenadas
    Exceptions        []Exception
    Holidays          []string         // fechas feriado nacional "YYYY-MM-DD", solo lectura
    OnWeeklyChange    func(WeeklyRow)   // fila editada (toggle/entrada/salida/colación)
    OnExceptionAdd    func(Exception)   // alta desde el panel (ID == "")
    OnExceptionRemove func(id string)
}

func (e *ScheduleEditor) Init(ctx dom.Ctx)
func (e *ScheduleEditor) Render() *dom.Element
```

## Comportamiento

### Plantilla semanal (`.scheduleeditor__week`)

- 7 filas, una por `WeeklyRow` en `Week` (el host garantiza 7, orden Dom→Sáb).
  Si `len(Week) != 7`: renderizar las que haya + `dom.Log` de dev-warning, sin
  panic (regla harness: lo que el compilador no caza cae a warning).
- Cada fila:
  - **Toggle activo** — `input[type=checkbox]`. Al cambiar: set `Active`,
    invocar `OnWeeklyChange(row)`. Fila inactiva: horas atenuadas por CSS
    (`[data-active="false"]`), pero **editables** (poner horas antes de activar
    es válido — como el legado "paso 1: elegir horas").
  - **Entrada / Salida / Colación desde / Colación hasta** — 4 `<select>`.
    Opciones cada 15 min de 06:00 a 22:00 (rango fijo del componente,
    documentado en el doc del paquete). Valor mostrado `HH:MM`, valor real
    minutos int. Colación vacía en ambos ⇒ `BreakStart=BreakFinish=0`.
  - Al cambiar cualquier select: recalcular la fila, `OnWeeklyChange(row)`.
- **Validación de dev-warning (no bloqueante; NO llama a `OnWeeklyChange`):**
  `WorkStart < WorkFinish`; con colación,
  `WorkStart <= BreakStart < BreakFinish <= WorkFinish`. Fila inválida: marca
  CSS `var(--color-error)` + `dom.Log`.

### Panel de excepciones (`.scheduleeditor__exceptions`)

Reutiliza `components/calendarslider` (API real, verificada):

```go
&calendarslider.CalendarSlider{
    NumMonths:  3,
    Holidays:   toCalHolidays(e.Holidays),          // []calendarslider.Holiday{Date,Name}
    Occupation: occupationFromExceptions(e.Exceptions), // []calendarslider.OccupationDay{Date,Percent}
    Selected:   e.sel,                               // *dom.SignalString
    OnSelect:   func(date string) { e.sel.Set(date) },
}
```

- `toCalHolidays` mapea `[]string` → `[]calendarslider.Holiday{Date: s, Name: "Feriado"}`.
- `occupationFromExceptions`: un `OccupationDay{Date, Percent}` por fecha con
  excepción — `Percent` 100 para HOLIDAY/BLOCKED, 50 para SPECIAL_HOURS. Solo
  sirve para que `calendarslider` haga el día seleccionable y lo marque
  (su regla: día con `Occupation` = clicable).
- Al elegir un día (`e.sel` != "") → formulario inline de alta
  (`.scheduleeditor__exc-form`, visibilidad por `e.sel != ""` bindeada, sin
  reconstruir el árbol):
  - `Date` prellenado con `e.sel.Get()` (solo lectura).
  - `Type` — 3 radios. Etiquetas visibles: el componente es librería → renderiza
    la palabra canónica inglesa vía `fmt/lang` `lang.Translate("Closed")` /
    `"Special hours"` / `"Blocked"` y registra **nada** (el diccionario lo pone
    la app — `layout/AGENTS.md` "Translatable messages"). Documentar estas 3
    claves + los 7 nombres de día en el doc del paquete / `README`.
  - `Type == SPECIAL_HOURS` o `BLOCKED`: dos `<select>` de hora (desde/hasta),
    mismo rango 06:00–22:00. HOLIDAY los oculta (bind sobre `e.excType`).
  - `Notes` — `input[type=text]` opcional.
  - Botón "Agregar" → `OnExceptionAdd(Exception{ID: "", Date, Type, StartMin, EndMin, Notes})`;
    limpia el form (reset `e.sel` a "").
- Lista de excepciones vigentes (`.scheduleeditor__exc-list`), orden fecha asc:
  fecha + etiqueta de tipo + horas si aplica + notas + "Quitar" →
  `OnExceptionRemove(id)`. Las de `Holidays` van solo-lectura, sin "Quitar",
  con marca visual distinta.

### Signals internas (no exportadas)

- `sel *dom.SignalString` — fecha elegida ("" = form oculto). Construida en `Init`.
- `excType *dom.SignalString` — tipo elegido (controla visibilidad de los
  `<select>` de hora).
- `Week`/`Exceptions`/`Holidays` de campo son estado **inicial**: el host
  persiste y remonta con datos frescos (mismo modelo que `targethour`/`crudview`
  — el componente no es la fuente de verdad). No mantener copia mutable interna
  más allá del render.

### CSS-first

- Tokens sin fallback: `var(--color-primary)`, `var(--color-error)`,
  `var(--mag-pri)`, etc. Sin `:root` en el `.css`.
- Atenuado de fila inactiva y visibilidad del form: CSS + `hidden`/`[data-*]`
  toggled desde el handler; no reconstruir el árbol.

## Tests P1

`scheduleeditor_test.go` (backend, `RenderHTML()`):
- `TestWeek_RendersSevenRows` / `TestWeek_InactiveRowMarked` (`data-active="false"`).
- `TestWeek_HourOptionsRange` — cada `<select>` con opciones 06:00..22:00 c/15m.
- `TestExceptions_ListSorted` — 3 excepciones desordenadas → render ordenado.
- `TestExceptions_HolidayReadonly` — fecha en `Holidays` → sin "Quitar".
- `TestSpecialHoursShowsTimeSelects` / `TestHolidayHidesTimeSelects` (según `excType`).
- Callbacks (dobles que capturan el último valor): cambio de select en la fila
  Lunes → `OnWeeklyChange` con `DayOfWeek==1` y minutos correctos; "Agregar" →
  `OnExceptionAdd` con `ID==""`.

`scheduleeditor_ui_wasm_test.go` (`//go:build wasm`): montar, clic en un día del
calendario → el form de alta se hace visible; submit → callback.

---

# Parte 2 — `components/targethour` gana `FreeSlots`

## Problema

`targethour.TargetHour` hoy tiene solo `Selected`, `OnSelect`, `StatusOf`
(`targethour.go:53`). **No** existe `FreeSlots`. La Etapa G necesita mostrar
huecos horarios reservables (derivados de `list_availability`) como filas
clicables junto a las reservas existentes.

## Cambio

En `targethour.go`, agregar al struct:

```go
// FreeSlots son horas "HH:MM" reservables (sin reserva). Se renderizan como
// filas ligeras al final de la lista, visualmente distintas de un Item real
// (sin estado, con un "+" o marco punteado). Opcional: nil = ninguna.
FreeSlots []string
// OnPickFree se invoca al hacer clic en un hueco libre, con su "HH:MM".
OnPickFree func(hhmm string)
```

- `Render()` / la construcción de filas: tras las filas de `items`, emitir una
  fila por cada `FreeSlots[i]` con clase `targethour__free` (o equivalente),
  `On("click", ...)` → `OnPickFree(hhmm)`.
- No participan de `listselect` (no son seleccionables para borrar/editar): son
  acciones de "reservar esta hora".
- CSS en `targethour/css.go` (`RenderSheet`/`RenderCSS` según el patrón del
  fichero): estilo `targethour__free` — atenuado, cursor pointer, marca de
  "disponible". Tokens sin fallback.
- Retrocompatible: `FreeSlots` nil ⇒ HTML idéntico a hoy.

## Tests P2

- `TestFreeSlots_RenderedAfterItems` — 2 items + 3 `FreeSlots` → 5 filas, las 3
  últimas con la clase `free`.
- `TestFreeSlots_ClickCallsOnPickFree` (`//go:build wasm`) — clic en un hueco →
  `OnPickFree` con el "HH:MM" correcto.
- `TestFreeSlots_NilNoRegression` — `FreeSlots` nil → HTML sin filas `free`.

---

## Criterios de aceptación (ambas partes)

- `gotest ./...` verde en `components`.
- `GOOS=js GOARCH=wasm go build ./...` OK.
- `go list -deps ./scheduleeditor/ | grep webtyp/svg/sprite` → vacío (SVG solo en
  `svg.go` con `//go:build !wasm`).
- `README.md` de `components` indexa `scheduleeditor`; `docs/CATALOG.md` con su
  entrada (formato de las existentes); si `docs/ARCHITECTURE.md` lista
  componentes, incluirlo. `targethour` doc/README menciona `FreeSlots`.
- `layout/docs/DICTIONARY.md` NO se toca desde aquí (es de `layout`), pero el
  doc de `scheduleeditor` lista las claves de traducción que introduce, para que
  la Etapa I las copie ahí.

## Fuera de alcance

- Múltiples bloques por día (master O2) — `WeeklyRow` queda como struct (no 4
  ints sueltos en la firma del callback) para no cerrar esa puerta.
- Persistencia / red — es del host (Etapa C/D).
- Cálculo de feriados o de huecos — el componente los recibe ya calculados.
