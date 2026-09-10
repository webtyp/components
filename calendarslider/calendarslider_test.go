//go:build !wasm

package calendarslider

import (
	"strings"
	"testing"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	"webtyp.com/widget"
)

// La aritmética pura de fechas (bisiestos, días por mes, día de la semana,
// suma de meses, claves YYYY-MM) vive y se prueba en webtyp.com/date
// — reutilizable fuera de este componente. Ver TestBuildMonthPadsToSixWeeks
// más abajo para la única cobertura que sigue siendo de calendarslider: cómo
// usa esa aritmética para armar la grilla.

// TestBuildMonthAgosto2026 fija la geometría del mes de agosto 2026: comienza
// sábado (5 huecos iniciales), tiene 31 días y las celdas finales quedan en la
// última semana parcial (1 día). El mes lleva: fila de días de la semana +
// 6 semanas + fila de navegación (prev + etiqueta + next en una sola fila).
func TestBuildMonthAgosto2026(t *testing.T) {
	c := &CalendarSlider{today: "2026-08-11"}
	m := c.buildMonth(2026, 8, "2026-07", "2026-09", false, false)
	if m == nil {
		t.Fatal("buildMonth returned nil")
	}
	children := m.Children()
	if len(children) != 8 {
		t.Fatalf("agosto 2026 debería tener fila de navegación + fila de días + 6 semanas = 8 hijos, tiene %d", len(children))
	}
	// La navegación abre la tarjeta: la etiqueta nombra la grilla que viene
	// abajo, así que se lee antes que los días.
	if !Contains(children[0].String(), "August 2026") {
		t.Errorf("el primer hijo debería ser la fila de navegación con la etiqueta en el medio: debería decir 'August 2026', dice %s", children[0].String())
	}
	if !Contains(children[0].String(), "<button") || !Contains(children[0].String(), "data-target='2026-07'") {
		t.Errorf("el botón anterior debería llevar data-target a julio, dice %s", children[0].String())
	}
	if !Contains(children[0].String(), "<button") || !Contains(children[0].String(), "data-target='2026-09'") {
		t.Errorf("el botón siguiente debería llevar data-target a septiembre, dice %s", children[0].String())
	}
	if children[1].String() != c.buildWeekdayRow().String() {
		t.Error("el segundo hijo del mes debería ser la fila de días de la semana")
	}
}

// TestBuildMonthAlwaysHasBothLinks cubre buildMonth en sí: siempre arma los
// dos enlaces con lo que se le pasa, sin importar si el llamador (Render)
// los está usando como vecino real o como vuelta del bucle. Quien decide el
// bucle infinito es Render, no buildMonth — ver TestRenderWrapsAround.
func TestBuildMonthAlwaysHasBothLinks(t *testing.T) {
	c := &CalendarSlider{today: "2026-08-11"}
	m := c.buildMonth(2026, 8, "2026-07", "2026-09", false, false).String()
	if !Contains(m, "Previous month") || !Contains(m, "Next month") {
		t.Error("buildMonth debería incluir siempre ambos enlaces")
	}
}

// TestRenderWrapsAround cubre el bucle infinito: el ‹ del primer mes de la
// tira apunta al último y el › del último apunta al primero — igual que el
// deslizador original, para volver al inicio sin recorrer los N meses.
func TestRenderWrapsAround(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08", NumMonths: 3}
	c.Init(nil)
	c.today = "2026-08-11"
	htmlOut := c.Render().String()

	if got := strings.Count(htmlOut, "calendarslider__prev"); got != 3 {
		t.Errorf("cada uno de los 3 meses debería tener un enlace 'prev', hay %d", got)
	}
	if got := strings.Count(htmlOut, "calendarslider__next"); got != 3 {
		t.Errorf("cada uno de los 3 meses debería tener un enlace 'next', hay %d", got)
	}

	// Agosto (primero) enlaza hacia atrás con octubre (último) — el bucle.
	augMonth := c.buildMonth(2026, 8, "2026-10", "2026-09", true, false).String()
	if !Contains(augMonth, "data-target='2026-10'") {
		t.Errorf("el 'prev' de agosto (primero) debería envolver a octubre (último), dice %s", augMonth)
	}
	// Octubre (último) enlaza hacia adelante con agosto (primero) — el bucle.
	octMonth := c.buildMonth(2026, 10, "2026-09", "2026-08", false, true).String()
	if !Contains(octMonth, "data-target='2026-08'") {
		t.Errorf("el 'next' de octubre (último) debería envolver a agosto (primero), dice %s", octMonth)
	}
}

// weeksPerMonth es siempre 6, ocupe o no el mes real sus 6 filas: sin este
// piso fijo, un mes de menos filas deja una tarjeta más baja y la etiqueta
// del mes (y el ‹ › que se ancla a la tarjeta) saltan verticalmente al
// deslizar entre meses. Febrero de 2021 empieza lunes y tiene 28 días — 4
// filas reales — el caso más corto posible.
func TestBuildMonthPadsToSixWeeks(t *testing.T) {
	c := &CalendarSlider{today: "2021-02-11"}
	m := c.buildMonth(2021, 2, "2021-01", "2021-03", false, false)
	children := m.Children()

	if len(children) != 8 {
		t.Fatalf("febrero 2021 debería tener 8 hijos (igual que cualquier mes), tiene %d", len(children))
	}

	for _, idx := range []int{6, 7} {
		week := children[idx]
		if !Contains(week.String(), "aria-hidden") {
			t.Errorf("la semana de relleno %d debería llevar aria-hidden", idx)
		}
		if Contains(week.String(), "data-date") {
			t.Errorf("la semana de relleno %d no debería contener ningún día real", idx)
		}
	}

	if !Contains(children[0].String(), "February 2021") {
		t.Errorf("la etiqueta debería abrir la tarjeta (posición 0), dice %s", children[0].String())
	}
}

// TestBuildMonthCells valida el contenido de las celdas: huecos iniciales,
// marcadores de hoy/feriado/domingo, ocupación y seleccionabilidad.
func TestBuildMonthCells(t *testing.T) {
	c := &CalendarSlider{
		today:      "2026-08-11",
		Holidays:   []Holiday{{Date: "2026-08-15", Name: "Asunción de la Virgen"}},
		Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 60}, {Date: "2026-08-02", Percent: 30}},
	}
	firstWeek := c.buildMonth(2026, 8, "2026-07", "2026-09", false, false).Children()[2]
	cells := firstWeek.Children()
	if len(cells) != 7 {
		t.Fatalf("primera semana debería tener 7 celdas, tiene %d", len(cells))
	}
	for i := 0; i < 5; i++ {
		if !Contains(cells[i].String(), "aria-hidden") {
			t.Errorf("celda hueco %d debería ser una celda vacía aria-hidden", i)
		}
	}

	day1 := cells[5]
	if !Contains(day1.String(), "data-date='2026-08-01'") {
		t.Error("día 1 debería caer en la columna 6")
	}
	if !Contains(day1.String(), "day-off") || Contains(day1.String(), "day-red") {
		t.Error("1 de agosto (sábado) sin ocupación ni feriado debería ser un día normal (day-off), el rojo es solo domingo/feriado")
	}
	if Contains(day1.String(), "day-selectable") || Contains(day1.String(), "data-use") {
		t.Error("un sábado sin ocupación no debería ser seleccionable")
	}

	day2 := cells[6]
	if !Contains(day2.String(), "day-selectable") {
		t.Error("domingo 2 con ocupación debería ser seleccionable")
	}
	if Contains(day2.String(), "day-red") {
		t.Error("la ocupación debería ganarle al domingo")
	}
	if !Contains(day2.String(), "data-use='30'") || !Contains(day2.String(), "--meter-fill:30%") {
		t.Error("la barra de ocupación debería llevar el porcentaje")
	}

	day11 := c.buildDay(2026, 8, 11)
	if !Contains(day11.String(), "day-today") {
		t.Error("el día de hoy debería marcarse con day-today")
	}
	if !Contains(day11.String(), "day-selectable") {
		t.Error("el día de hoy con ocupación debería ser seleccionable")
	}
	if !Contains(day11.String(), "title='Today'") {
		t.Error("el día de hoy debería llevar title 'Today'")
	}

	day15 := c.buildDay(2026, 8, 15)
	if !Contains(day15.String(), "day-red") {
		t.Error("el feriado debería marcarse rojo")
	}
	if !Contains(day15.String(), "Asunción de la Virgen") {
		t.Error("el feriado debería llevar el nombre en el title")
	}
	if Contains(day15.String(), "day-selectable") {
		t.Error("un feriado no debería ser seleccionable")
	}

	day14 := c.buildDay(2026, 8, 14)
	if !Contains(day14.String(), "day-off") {
		t.Error("día hábil sin ocupación debería marcarse como inactivo")
	}
	if Contains(day14.String(), "day-selectable") {
		t.Error("día sin ocupación no debería ser seleccionable")
	}
}

func TestClampOccupation(t *testing.T) {
	c := &CalendarSlider{Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 150}, {Date: "2026-08-12", Percent: -5}}}
	if !Contains(c.buildDay(2026, 8, 11).String(), "data-use='100'") {
		t.Error("ocupación sobre 100 debería recortarse a 100")
	}
	if !Contains(c.buildDay(2026, 8, 12).String(), "data-use='0'") {
		t.Error("ocupación negativa debería recortarse a 0")
	}
}

func TestRenderStructure(t *testing.T) {
	c := &CalendarSlider{
		Start:      "2026-08",
		NumMonths:  3,
		Holidays:   []Holiday{{Date: "2026-08-15", Name: "Asunción"}},
		Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 60}},
	}
	c.Init(nil)
	c.today = "2026-08-11"

	htmlOut := c.Render().String()

	for _, key := range []string{"2026-08", "2026-09", "2026-10"} {
		if !strings.Contains(htmlOut, "data-month='"+key+"'") {
			t.Errorf("la tira debería incluir el mes %q, no aparece en:\n%s", key, htmlOut)
		}
	}

	if !Contains(htmlOut, "data-target='2026-09'") {
		t.Error("agosto (Start) debería apuntar a septiembre como mes siguiente")
	}

	cssOut := c.RenderCSS().String()
	if strings.Contains(cssOut, "position: absolute;") {
		t.Error("nada debería flotar con position:absolute (ni flechas ni chip)")
	}
	if strings.Contains(cssOut, "inset-block: 0;") {
		t.Error("las flechas no deberían llevar EdgeStrip (inset-block: 0) en ningún modo")
	}

	// Claves canónicas en inglés: la librería no fija idioma, lo traduce el
	// diccionario de la app. Y el orden lo manda date.FirstWeekday, que es
	// lunes por defecto — el domingo va último, no primero.
	if !Contains(htmlOut, "Mon") || !Contains(htmlOut, "Sun") {
		t.Error("la fila de días de la semana debería estar dentro de cada mes")
	}
	if mon, sun := Index(htmlOut, "Mon"), Index(htmlOut, "Sun"); mon < 0 || sun < 0 || mon > sun {
		t.Errorf("la semana debe empezar en lunes y terminar en domingo (Mon=%d, Sun=%d)", mon, sun)
	}
	if !strings.Contains(htmlOut, "calendarslider__week-row") {
		t.Error("las semanas deberían llevar la clase de grilla de 7 columnas")
	}
	if !Contains(htmlOut, "August 2026") {
		t.Error("cada mes debería llevar su etiqueta (mes + año)")
	}

	c.Selected.Set("2026-08-11")
	htmlOut = c.Render().String()
	if !strings.Contains(htmlOut, "data-selected='true'") {
		t.Error("las celdas deberían llevar el estado data-selected")
	}
}

func TestPairMarkupAndStylesheet(t *testing.T) {
	// Un día por cada escalón de la rampa de ocupación: la aserción es que
	// NINGUNA clase de la hoja quede muerta, así que el fixture tiene que
	// ejercitar las tres, no solo la que toque.
	c := &CalendarSlider{Start: "2026-08", Holidays: []Holiday{{Date: "2026-08-15", Name: "x"}}, Occupation: []OccupationDay{
		{Date: "2026-08-11", Percent: 10},
		{Date: "2026-08-12", Percent: 60},
		{Date: "2026-08-13", Percent: 95},
	}}
	c.Init(nil)
	c.today = "2026-08-11"
	htmlOut := c.Render().String()
	cssOut := c.RenderCSS().String()

	extractClasses := func(hay, prefix string) map[string]bool {
		out := make(map[string]bool)
		rest := hay
		for {
			idx := strings.Index(rest, prefix)
			if idx < 0 {
				break
			}
			end := idx
			for end < len(rest) && rest[end] != ' ' && rest[end] != '{' && rest[end] != ',' && rest[end] != '}' && rest[end] != '\n' && rest[end] != '\r' && rest[end] != '\t' && rest[end] != ':' && rest[end] != '[' {
				end++
			}
			out[rest[idx:end]] = true
			rest = rest[end:]
		}
		return out
	}

	for cls := range extractClasses(cssOut, "calendarslider__") {
		if !strings.Contains(htmlOut, cls) {
			t.Errorf("clase CSS %q no existe en el HTML renderizado", cls)
		}
	}
}

func TestNumMonthsClampsToMax(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08", NumMonths: 20}
	c.Init(nil)
	htmlOut := c.Render().String()

	if got := strings.Count(htmlOut, "data-month='"); got != maxMonths {
		t.Fatalf("NumMonths=20 debería recortarse a %d meses, la tira tiene %d", maxMonths, got)
	}
	if !Contains(htmlOut, "data-month='2026-08'") {
		t.Error("el primer mes de una tira de 12 empezando en agosto 2026 debería ser agosto 2026 (Start)")
	}
	if !Contains(htmlOut, "data-month='2027-07'") {
		t.Error("el último mes de una tira de 12 empezando en agosto 2026 debería ser julio 2027")
	}
}

func TestNumMonthsDefaultsToThree(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	htmlOut := c.Render().String()
	if got := strings.Count(htmlOut, "data-month='"); got != 3 {
		t.Errorf("NumMonths sin especificar debería dar 3 meses, la tira tiene %d", got)
	}
}

func TestCalendarSlider_SatisfiesFilterable(t *testing.T) {
	var c widget.Filterable = &CalendarSlider{}
	got := ""
	c.OnFilterChange(func(term string) { got = term })
	if got != "" {
		t.Fatalf("sink must not fire on registration, got %q", got)
	}
}

func TestNavRowAlwaysVisible(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	cssOut := c.RenderCSS().String()

	for _, sel := range []string{
		".calendarslider__month:hover .calendarslider__month-nav",
		".calendarslider__month:focus-within .calendarslider__month-nav",
	} {
		if strings.Contains(cssOut, sel) {
			t.Errorf("la fila de navegación no debería ocultarse nunca (%s presente)", sel)
		}
	}
	if strings.Contains(cssOut, "inset-block: 0;") {
		t.Errorf("las flechas no deberían llevar EdgeStrip en ningún modo")
	}
	if !strings.Contains(cssOut, ".calendarslider__month-nav") {
		t.Errorf("la fila de navegación debería existir en la hoja de estilos")
	}
}

func TestDaysFillTheirTrack(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	cssOut := c.RenderCSS().String()

	idx := strings.Index(cssOut, ".calendarslider__day {")
	if idx < 0 {
		t.Fatal("la regla base del día no aparece")
	}
	rule := cssOut[idx:]
	if end := strings.Index(rule, "}"); end >= 0 {
		rule = rule[:end]
	}
	if !strings.Contains(rule, "aspect-ratio") {
		t.Errorf("el día debería dimensionarse por aspect-ratio (cuadrado de su columna), dice:\n%s", rule)
	}
	if strings.Contains(rule, "width: 1.5em") || strings.Contains(rule, "height: 1.5em") {
		t.Errorf("el día no debería llevar caja fija IconBox, dice:\n%s", rule)
	}
	if media := strings.Index(cssOut, "@media"); media >= 0 {
		if strings.Contains(cssOut[media:], ".calendarslider__day {") {
			t.Errorf("el día no debería tener regla móvil propia (el aspect cubre todo)")
		}
	}
	foundGap := false
	rest := cssOut
	for {
		i := strings.Index(rest, ".calendarslider__week-row {")
		if i < 0 {
			break
		}
		block := rest[i:]
		if end := strings.Index(block, "}"); end >= 0 {
			block = block[:end]
		}
		if strings.Contains(block, "var(--space-2") {
			foundGap = true
			break
		}
		rest = rest[i+1:]
	}
	if !foundGap {
		t.Errorf("la grilla debería llevar gap Space2 uniforme")
	}
}

func TestMonthNavMatchesEcosystemSize(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	cssOut := c.RenderCSS().String()

	ruleBlocks := func(sel string) []string {
		var out []string
		rest := cssOut
		for {
			idx := strings.Index(rest, sel+" {")
			if idx < 0 {
				break
			}
			block := rest[idx:]
			if end := strings.Index(block, "}"); end >= 0 {
				block = block[:end]
			}
			out = append(out, block)
			rest = rest[idx+len(sel):]
		}
		if len(out) == 0 {
			t.Fatalf("la regla %s no aparece", sel)
		}
		return out
	}
	ruleHas := func(sel, want string) bool {
		for _, b := range ruleBlocks(sel) {
			if strings.Contains(b, want) {
				return true
			}
		}
		return false
	}
	for _, sel := range []string{".calendarslider__prev", ".calendarslider__next"} {
		if !ruleHas(sel, "width: 2.5em") || !ruleHas(sel, "height: 2.5em") {
			t.Errorf("%s debería ser caja exacta 50px (IconBox Lg a TextXl)", sel)
		}
		if !ruleHas(sel, "font-size: var(--text-lg") {
			t.Errorf("%s debería escalar con TextXl (fuente única con IconBox)", sel)
		}
	}
	if !ruleHas(".calendarslider__month-name", "font-size: var(--text-sm") {
		t.Errorf("la etiqueta del mes debería ir en TextSm como el chip")
	}
}

func TestCollapseWorksEverywhere(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	cssOut := c.RenderCSS().String()

	stripShown := ".calendarslider__strip[data-current=\"true\"]"
	if !strings.Contains(cssOut, stripShown) {
		t.Fatalf("la tira debería mostrarse por estado (%s), no aparece", stripShown)
	}
	if !strings.Contains(cssOut, "var(--duration-slow") {
		t.Errorf("el reveal debería correr en MotionSlow (400ms)")
	}
	if idx, media := strings.Index(cssOut, stripShown), strings.Index(cssOut, "@media"); idx > media && media >= 0 {
		t.Errorf("la regla %s debería ser top-level (todos los viewports), no dentro de un @media", stripShown)
	}

	base := ".calendarslider__collapsed {"
	idx := strings.Index(cssOut, base)
	if idx < 0 {
		t.Fatalf("la base del chip (%s) no aparece", base)
	}
	rule := cssOut[idx:]
	if end := strings.Index(rule, "}"); end >= 0 {
		rule = rule[:end]
	}
	if strings.Contains(rule, "display: none;") {
		t.Errorf("el chip no debería ocultarse en su base (es el toggle en ambos modos), dice:\n%s", rule)
	}
}

func TestCollapsedChipIsTopField(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	cssOut := c.RenderCSS().String()

	kids := c.Render().Children()
	if len(kids) != 2 {
		t.Fatalf("la raíz debería tener campo + tira, tiene %d hijos", len(kids))
	}
	// El campo es el DISPARADOR: va arriba, y la tira del mes cuelga de él.
	// Antes quedaba debajo del panel que abría.
	if !Contains(kids[0].String(), "calendarslider__collapsed") {
		t.Errorf("el campo debería ser el primer hijo, dice:\n%s", kids[0].String())
	}
	if !Contains(kids[1].String(), "calendarslider__strip") {
		t.Errorf("la tira debería ser el segundo hijo, dice:\n%s", kids[1].String())
	}
	for _, banned := range []string{
		"inset-block-end:",
		"inset-inline-start:",
	} {
		if strings.Contains(cssOut, banned) {
			t.Errorf("el chip no debería anclarse (Docked eliminado): %q presente", banned)
		}
	}
}

func TestMultiSelectToggle(t *testing.T) {
	selMany := NewString("2026-08-11")

	c := &CalendarSlider{
		today:        "2026-08-11",
		Holidays:     []Holiday{{Date: "2026-08-15", Name: "Asunción de la Virgen"}},
		Occupation:   []OccupationDay{{Date: "2026-08-11", Percent: 50}, {Date: "2026-08-12", Percent: 50}},
		SelectedMany: selMany,
	}
	c.Init(nil)

	// Initial derivation: the seeded key is selected, its neighbour is not.
	if !strings.Contains(c.buildDay(2026, 8, 11).String(), "data-selected='true'") {
		t.Error("2026-08-11 should be selected initially in multi-select mode")
	}
	if strings.Contains(c.buildDay(2026, 8, 12).String(), "data-selected='true'") {
		t.Error("2026-08-12 should not be selected initially")
	}

	// Occupation/Holidays still paint in multi-select mode.
	if !strings.Contains(c.buildDay(2026, 8, 15).String(), "day-red") {
		t.Error("a holiday should still paint red with SelectedMany set")
	}
	if strings.Contains(c.buildDay(2026, 8, 15).String(), "day-selectable") {
		t.Error("a holiday must not be selectable in multi-select mode")
	}

	// Toggle ON: 2026-08-12 joins the set, 2026-08-11 stays.
	next, on := toggleWord(selMany.Get(), "2026-08-12")
	if !on {
		t.Fatal("toggling an unselected date should report it now selected")
	}
	selMany.Set(next)
	if !containsWord(selMany.Get(), "2026-08-11") || !containsWord(selMany.Get(), "2026-08-12") {
		t.Fatalf("both dates should be in the set after toggle-on, got %q", selMany.Get())
	}
	if !strings.Contains(c.buildDay(2026, 8, 12).String(), "data-selected='true'") {
		t.Error("2026-08-12 should render selected after toggle-on")
	}

	// Toggle OFF: 2026-08-11 leaves the set, 2026-08-12 stays.
	next, on = toggleWord(selMany.Get(), "2026-08-11")
	if on {
		t.Fatal("toggling a selected date should report it now unselected")
	}
	selMany.Set(next)
	if containsWord(selMany.Get(), "2026-08-11") {
		t.Errorf("2026-08-11 should be gone after toggle-off, got %q", selMany.Get())
	}
	if strings.Contains(c.buildDay(2026, 8, 11).String(), "data-selected='true'") {
		t.Error("2026-08-11 should not render selected after toggle-off")
	}
}

// TestSingleSelectUnchangedBySelectedMany guards the coexistence rule: with
// SelectedMany nil, selection is derived from Selected exactly as before, so
// every existing consumer (targethour, reservation, the exceptions panel) is
// untouched by the multi-select addition.
func TestSingleSelectUnchangedBySelectedMany(t *testing.T) {
	c := &CalendarSlider{
		today:      "2026-08-11",
		Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 50}, {Date: "2026-08-12", Percent: 50}},
		Selected:   NewString("2026-08-12"),
	}
	c.Init(nil)

	if !strings.Contains(c.buildDay(2026, 8, 12).String(), "data-selected='true'") {
		t.Error("Selected date should render selected when SelectedMany is nil")
	}
	if strings.Contains(c.buildDay(2026, 8, 11).String(), "data-selected='true'") {
		t.Error("a non-Selected date must not render selected in single-select mode")
	}
}
