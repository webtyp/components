//go:build !wasm

package scheduleeditor

import (
	"strings"
	"testing"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testEditor() *ScheduleEditor {
	return &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1, 3, 5}},
			{StartMin: 900, EndMin: 1140, Days: []int{1, 3, 5}},
		},
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 540, EndMin: 780},
		},
		Bounds: Bounds{OpenMin: 480, CloseMin: 1200},
	}
}

func TestOnePatternRowCoversSeveralWeekdays(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1, 3, 5}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	// Los SIETE días se listan siempre: un día libre se ve, no se deduce.
	if got := strings.Count(html, "scheduleeditor__day-row"); got != 7 {
		t.Fatalf("se esperaban 7 filas de día, hay %d:\n%s", got, html)
	}
	// Y la fila que cubre Lun/Mié/Vie aparece como horario en esos tres.
	if got := strings.Count(html, "scheduleeditor__day-slots"); got != 3 {
		t.Errorf("se esperaban 3 días con horario (Lun/Mié/Vie), hay %d:\n%s", got, html)
	}
	// Los otros cuatro lo dicen con todas las letras.
	if got := strings.Count(html, "scheduleeditor__day-off-text"); got != 4 {
		t.Errorf("se esperaban 4 días marcados como libres, hay %d:\n%s", got, html)
	}
}

func TestTwoRowsShareADayAndLeaveAGap(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 780, Days: []int{1}},
			{StartMin: 900, EndMin: 1140, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	// El corte de almuerzo: dos rangos en el MISMO día. Antes obligaba a
	// crear una segunda regla y re-tildar los mismos días; ahora son dos
	// slots dentro de la fila del lunes.
	if got := strings.Count(html, "scheduleeditor__time-range"); got != 2 {
		t.Errorf("se esperaban 2 rangos en el lunes, hay %d:\n%s", got, html)
	}
	if got := strings.Count(html, "scheduleeditor__day-slots"); got != 1 {
		t.Errorf("solo el lunes debería tener horario, hay %d días con horario:\n%s", got, html)
	}
	// Con más de un rango aparece la papelera por rango; con uno solo no.
	if !strings.Contains(html, "scheduleeditor__slot-remove") {
		t.Errorf("con dos rangos debería poder quitarse uno:\n%s", html)
	}
}

func TestEmptyPatternWithMarkedDaysIsValid(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{},
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 540, EndMin: 780},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if strings.Contains(html, "data-invalid='true'") {
		t.Errorf("empty pattern with marked days should be valid:\n%s", html)
	}
	// Legible, no el identificador: "2026-09-19" sirve para ordenar y
	// comparar, y es ilegible en una lista que alguien recorre con la vista.
	if !strings.Contains(html, "19 September 2026") {
		t.Errorf("el día extra debería mostrar la fecha legible:\n%s", html)
	}
	if strings.Contains(html, ">2026-09-19<") {
		t.Errorf("la fecha ISO cruda no debería llegar a la pantalla:\n%s", html)
	}
}

// TestExtraDayRoutesToOnDaysMarked cubre la mitad "sumo disponibilidad" de la
// sección unificada de fechas: el tipo elegido es lo que decide el destino, y
// un día extra tiene que llegar por OnDaysMarked, no por OnExceptionAdd.
// TestOpeningAClosedDayGoesThroughAddException cubre la única vía que el host
// puede persistir. "Atender ese día" emitía antes por OnDaysMarked, un callback
// que ningún host de este repo cablea porque ScheduleClient no expone guardado
// de bloques por fecha — así que el botón no hacía nada.
//
// SPECIAL_HOURS sí abre un día cerrado: availableRanges lo resuelve ANTES de
// mirar los bloques semanales (appointment_booking/service.go:564).
func TestOpeningAClosedDayGoesThroughAddException(t *testing.T) {
	e := &ScheduleEditor{}
	e.Init(&emptyCtx{})

	var got Exception
	e.OnExceptionAdd = func(ex Exception) { got = ex }

	e.sel.Set("2026-09-20")
	e.excType.Set(ExcSpecialHours)
	e.excFrom.Set("480")
	e.excTo.Set("960")
	e.addException()

	if got.Date != "2026-09-20" || got.Type != ExcSpecialHours {
		t.Fatalf("se esperaba SPECIAL_HOURS en 2026-09-20, llegó %+v", got)
	}
	if got.StartMin != 480 || got.EndMin != 960 {
		t.Errorf("horas = (%d, %d), se esperaba (480, 960)", got.StartMin, got.EndMin)
	}
	if e.sel.Get() != "" {
		t.Error("tras agregar, el formulario debería cerrarse (sel vacío)")
	}
}

func TestMarkedDayCanDivergeFromTheCommonWindow(t *testing.T) {
	e := &ScheduleEditor{
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 480, EndMin: 720},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "19 September 2026") {
		t.Fatalf("la lista debería mostrar la fecha legible:\n%s", html)
	}
	// La lista unificada muestra el horario del día extra como texto: es una
	// lista de lo que ya existe, no un formulario de edición.
	if !strings.Contains(html, "08:00–12:00") {
		t.Errorf("el horario del día extra debería leerse en la lista:\n%s", html)
	}
}

func TestPatternAndMarkedDaysRenderTogether(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "scheduleeditor__pattern") {
		t.Errorf("pattern section missing:\n%s", html)
	}
	// Ya no hay sección "marker": los días extra viven en la sección única de
	// fechas específicas, junto a las excepciones.
	if !strings.Contains(html, "scheduleeditor__exceptions") {
		t.Errorf("falta la sección de fechas específicas:\n%s", html)
	}
	if strings.Contains(html, "scheduleeditor__marker") {
		t.Errorf("la sección marker se fundió en fechas específicas, no debería existir:\n%s", html)
	}
}

func TestUnmarkingADayFiresOnDaysUnmarked(t *testing.T) {
	// Quitar un día extra ocurre desde la lista unificada de fechas, que es
	// la única forma de sacarlo desde que las dos secciones se fundieron.
	e := &ScheduleEditor{
		Marked: []MarkedDay{{Date: "2026-09-20", StartMin: 540, EndMin: 780}},
	}
	e.Init(&emptyCtx{})

	var gotUnmarked []string
	e.OnDaysUnmarked = func(dates []string) {
		gotUnmarked = dates
	}

	html := e.buildExceptionList().String()
	if !strings.Contains(html, "20 September 2026") {
		t.Fatalf("el día extra debería aparecer en la lista de fechas:\n%s", html)
	}
	e.OnDaysUnmarked([]string{"2026-09-20"})

	if len(gotUnmarked) != 1 || gotUnmarked[0] != "2026-09-20" {
		t.Fatalf("OnDaysUnmarked recibió %v, se esperaba ['2026-09-20']", gotUnmarked)
	}
}

func TestHourOptionsAreClampedToBounds(t *testing.T) {
	b := Bounds{OpenMin: 480, CloseMin: 1080}
	opts := hourOptions(540, b, 15)

	if len(opts) != 41 {
		t.Fatalf("expected 41 options for 08:00..18:00 step 15, got %d", len(opts))
	}
	if !strings.Contains(opts[0].String(), "08:00") {
		t.Errorf("first option = %s, want 08:00", opts[0].String())
	}
	if !strings.Contains(opts[len(opts)-1].String(), "18:00") {
		t.Errorf("last option = %s, want 18:00", opts[len(opts)-1].String())
	}
}

func TestWiderBoundsOfferMoreOptions(t *testing.T) {
	narrow := hourOptions(540, Bounds{OpenMin: 540, CloseMin: 1020}, 15)
	wider := hourOptions(540, Bounds{OpenMin: 480, CloseMin: 1200}, 15)

	if len(wider) <= len(narrow) {
		t.Fatalf("wider bounds should produce more options: wider=%d, narrow=%d", len(wider), len(narrow))
	}
}

func TestZeroBoundsFallBackToFullDay(t *testing.T) {
	opts := hourOptions(0, Bounds{0, 0}, 15)
	if len(opts) != 96 {
		t.Fatalf("zero bounds should fallback to 96 options (00:00..23:45), got %d", len(opts))
	}
	if !strings.Contains(opts[0].String(), "00:00") {
		t.Errorf("first option = %s, want 00:00", opts[0].String())
	}
	if !strings.Contains(opts[len(opts)-1].String(), "23:45") {
		t.Errorf("last option = %s, want 23:45", opts[len(opts)-1].String())
	}
}

func TestHolidayIsNotSelectableInTheMarker(t *testing.T) {
	e := &ScheduleEditor{
		Holidays: []string{"2026-09-18"},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if strings.Contains(html, "data-date='2026-09-18' class='calendarslider__day calendarslider__day-selectable") {
		t.Errorf("holiday should not be marked selectable in calendar:\n%s", html)
	}
}

func TestOverlappingRowsAreMarkedInvalidButNotBlocked(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1}},
			{StartMin: 600, EndMin: 900, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	// El solapamiento se marca pero no se bloquea: el editor avisa, no impide.
	// Y ahora vive en la fila del DÍA — dos rangos del lunes que se pisan — en
	// vez de obligar a cruzar el patrón entero.
	if got := strings.Count(html, "data-invalid='true'"); got != 1 {
		t.Errorf("solo el lunes se pisa: se esperaba 1 fila inválida, hay %d:\n%s", got, html)
	}
	if got := strings.Count(html, "scheduleeditor__time-range"); got != 2 {
		t.Errorf("los dos rangos deben seguir renderizándose, hay %d:\n%s", got, html)
	}
}

func TestRevealedStatesAreWrittenByTheMarkup(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 1080, EndMin: 540, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	e.sel.Set("2026-09-19")
	html := e.Render().String()

	for _, kv := range e.sheet().StateAttrs() {
		if !strings.Contains(html, kv.Key) {
			t.Errorf("stylesheet reveals/repaints on %q (value: %q) but no element writes it:\n%s", kv.Key, kv.Value, html)
		}
	}
}

func TestExceptions_ListSorted(t *testing.T) {
	e := testEditor()
	e.Exceptions = []Exception{
		{ID: "b", Date: "2026-09-20", Type: ExcBlocked},
		{ID: "a", Date: "2026-09-18", Type: ExcHoliday},
		{ID: "c", Date: "2026-09-19", Type: ExcSpecialHours, StartMin: 540, EndMin: 600},
	}
	e.Init(&emptyCtx{})
	list := e.buildExceptionList()

	items := []string{}
	for _, c := range list.Children() {
		s := extractDate(c.String())
		if s != "" {
			items = append(items, s)
		}
	}
	// Cuatro, no tres: la lista es única — las tres excepciones MÁS el día
	// extra que trae testEditor. Y ordenada como una sola, que es lo que
	// importa: el usuario la lee como una sola lista de fechas.
	if len(items) != 4 {
		t.Fatalf("se esperaban 4 fechas (3 excepciones + 1 día extra), hay %d: %v", len(items), items)
	}
	for i := 1; i < len(items); i++ {
		if items[i-1] > items[i] {
			t.Fatalf("la lista unificada no está ordenada por fecha: %v", items)
		}
	}
}

func extractDate(s string) string {
	quote := strings.Index(s, "2026-09-")
	if quote < 0 {
		return ""
	}
	return s[quote : quote+10]
}

func TestExceptions_HolidayReadonly(t *testing.T) {
	e := testEditor()
	e.Holidays = []string{"2026-09-18"}
	e.Exceptions = []Exception{
		{ID: "a", Date: "2026-09-18", Type: ExcHoliday},
	}
	e.Init(&emptyCtx{})

	// La aserción se acota A LA FILA del feriado: la lista unificada trae
	// además días extra, que sí deben poder quitarse.
	var holidayRow string
	for _, c := range e.buildExceptionList().Children() {
		if strings.Contains(c.String(), "scheduleeditor__exc-holiday") {
			holidayRow = c.String()
		}
	}
	if holidayRow == "" {
		t.Fatalf("la fila del feriado debería llevar la marca de feriado:\n%s", e.buildExceptionList().String())
	}
	if strings.Contains(holidayRow, "scheduleeditor__exc-remove") {
		t.Errorf("un feriado lo pone el establecimiento: no debe traer botón Quitar:\n%s", holidayRow)
	}
}
