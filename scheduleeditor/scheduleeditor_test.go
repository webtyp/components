//go:build !wasm

package scheduleeditor

import (
	"strings"
	"testing"

	"webtyp.com/date"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testEditor() *ScheduleEditor {
	return &ScheduleEditor{
		Week: []WeeklyRow{
			{Active: false, WorkStart: 0, WorkFinish: 0},
			{Active: true, WorkStart: 540, WorkFinish: 1020, BreakStart: 780, BreakFinish: 840},
			{Active: true, WorkStart: 540, WorkFinish: 1020},
			{Active: true, WorkStart: 540, WorkFinish: 1020},
			{Active: true, WorkStart: 540, WorkFinish: 1020},
			{Active: true, WorkStart: 540, WorkFinish: 1020},
			{Active: false, WorkStart: 0, WorkFinish: 0},
		},
	}
}

// La grilla semanal renderiza las 7 filas con su día y su toggle.
func TestWeek_RendersSevenRows(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.buildWeek().String()

	for _, dow := range []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"} {
		if !strings.Contains(html, dow) {
			t.Errorf("missing day %q in week:\n%s", dow, html)
		}
	}
	for _, want := range []string{"scheduleeditor__week-row", "scheduleeditor__toggle", "scheduleeditor__time"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in week:\n%s", want, html)
		}
	}
}

// Fila inactiva: data-active="false".
func TestWeek_InactiveRowMarked(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.buildWeekRow(0, e.Week[0]).String()

	if !strings.Contains(html, "data-active='false'") {
		t.Errorf("inactive row must carry data-active='false':\n%s", html)
	}
}

// Cada <select> de hora lleva opciones de 06:00 a 22:00 cada 15 minutos.
func TestWeek_HourOptionsRange(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})

	for _, row := range []int{1, 2} {
		html := e.buildWeekRow(row, e.Week[row]).String()
		if !strings.Contains(html, "06:00") || !strings.Contains(html, "22:00") {
			t.Errorf("row %d must offer the 06:00..22:00 range:\n%s", row, html)
		}
	}
}

// Las opciones son 06:00..22:00 en saltos de 15 min -> 65 opciones por select.
func TestWeek_HourOptionsCount(t *testing.T) {
	opts := hourOptions(540)
	if len(opts) != 65 {
		t.Fatalf("expected 65 options (06:00..22:00 each 15m), got %d", len(opts))
	}
	if opts[0].String() != `<option value='360'>06:00</option>` {
		t.Errorf("first option = %q, want 06:00/360", opts[0].String())
	}
}

// Excepciones: se renderizan ordenadas por fecha.
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
	if len(items) != 3 || items[0] != "2026-09-18" || items[1] != "2026-09-19" || items[2] != "2026-09-20" {
		t.Fatalf("exceptions not sorted by date: %v", items)
	}
}

func extractDate(s string) string {
	quote := strings.Index(s, "2026-09-")
	if quote < 0 {
		return ""
	}
	return s[quote : quote+10]
}

// Un día feriado (en Holidays) se muestra sin "Quitar".
func TestExceptions_HolidayReadonly(t *testing.T) {
	e := testEditor()
	e.Holidays = []string{"2026-09-18"}
	e.Exceptions = []Exception{
		{ID: "a", Date: "2026-09-18", Type: ExcHoliday},
	}
	e.Init(&emptyCtx{})

	html := e.buildExceptionList().String()
	if strings.Contains(html, "scheduleeditor__exc-remove") {
		t.Errorf("holiday exception must not render a Remove button:\n%s", html)
	}
	if !strings.Contains(html, "scheduleeditor__exc-holiday") {
		t.Errorf("holiday-exception row must carry the holiday mark:\n%s", html)
	}
}

// El formulario de alta: con excType SPECIAL_HOURS (default) muestra los
// selects de hora (la fila hours lleva data-open="true"); con HOLIDAY los
// oculta. El form en sí lleva data-open="true" mientras hay día elegido, así
// que el marcador del "horario visible" es que data-open aparezca DOS veces
// (form + hours).
func TestSpecialHoursShowsTimeSelects(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	e.excType.Set(ExcSpecialHours)
	e.sel.Set("2026-09-19")

	html := e.buildExceptionForm().String()
	if countOpen(html) != 2 {
		t.Errorf("SPECIAL_HOURS must keep the hours row open (2x data-open='true'):\n%s", html)
	}
}

func TestHolidayHidesTimeSelects(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	e.excType.Set(ExcHoliday)
	e.sel.Set("2026-09-19")

	html := e.buildExceptionForm().String()
	if countOpen(html) != 1 {
		t.Errorf("HOLIDAY must hide the hours row (exactly 1 data-open='true', the form's):\n%s", html)
	}
	if !strings.Contains(html, "scheduleeditor__exc-type") {
		t.Errorf("the type radio group must render:\n%s", html)
	}
}

func countOpen(html string) int {
	n := 0
	for i := 0; i+len("data-open='true'") <= len(html); i++ {
		if html[i:i+len("data-open='true'")] == "data-open='true'" {
			n++
		}
	}
	return n
}

// Callbacks: "Agregar" con un día elegido -> OnExceptionAdd con ID == "" y los
// datos del form; un cambio en la fila Lunes -> OnWeeklyChange con la fila.
func TestCallbacks_AddFiresOnExceptionAdd(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	e.sel.Set("2026-09-25")
	e.excType.Set(ExcBlocked)

	var got *Exception
	e.OnExceptionAdd = func(ex Exception) { got = &ex }

	e.addException()

	if got == nil {
		t.Fatal("OnExceptionAdd not called")
	}
	if got.ID != "" {
		t.Errorf("expected ID == \"\" on add, got %q", got.ID)
	}
	if got.Date != "2026-09-25" || got.Type != ExcBlocked {
		t.Errorf("unexpected exception: %+v", got)
	}
	if got.StartMin != 540 || got.EndMin != 1080 {
		t.Errorf("BLOCKED should carry the default window: %+v", got)
	}
}

// Un cambio de entrada en la fila Lunes (índice 1) recompone la fila completa
// y llama OnWeeklyChange con dayOfWeek==1 y los minutos editados.
func TestCallbacks_WeeklyChangeMonday(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})

	var gotDay int
	var got *WeeklyRow
	e.OnWeeklyChange = func(d int, r WeeklyRow) { gotDay, got = d, &r }

	// The weekly select's change handler recomputes from the ORIGINAL row with
	// one field set — exercising the same path the DOM wiring uses. We call the
	// component's action directly (the handler is a small wrapper over it).
	e.weeklyChange(1, func(r *WeeklyRow) { r.WorkStart = 600 })

	if got == nil {
		t.Fatal("OnWeeklyChange not called")
	}
	if gotDay != 1 {
		t.Errorf("dayOfWeek = %d, want 1 (Monday)", gotDay)
	}
	if got.WorkStart != 600 {
		t.Errorf("WorkStart = %d, want 600", got.WorkStart)
	}
	// Rest of the row preserved.
	if got.WorkFinish != 1020 || got.BreakStart != 780 || got.BreakFinish != 840 {
		t.Errorf("row not preserved: %+v", *got)
	}
}

// Every state the stylesheet reveals or repaints on must actually be written
// by the markup. The two halves live behind different build tags and nothing
// checks them: this is the loop widget/docs/DESIGN.md §17 says the consumer
// closes. The week carries one invalid row so data-invalid is written too.
func TestRevealedStatesAreWrittenByTheMarkup(t *testing.T) {
	week := make([]WeeklyRow, 7)
	week[1] = WeeklyRow{Active: true, WorkStart: 600, WorkFinish: 480} // invalid window
	e := &ScheduleEditor{Week: week}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	for _, kv := range e.sheet().StateAttrs() {
		if !strings.Contains(html, kv.Key) {
			t.Errorf("stylesheet reveals/repaints on %q but no element writes it:\n%s", kv.Key, html)
		}
	}
}

// The seven rows render the seven day names in order. The bug this replaces:
// the host left the day at its zero value on unconfigured days and four rows
// rendered "Sunday".
func TestWeekRendersSevenDistinctDaysInOrder(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	for i := 0; i < 7; i++ {
		want := date.WeekdayName(i)
		if !strings.Contains(html, want) {
			t.Errorf("row %d: missing day name %q:\n%s", i, want, html)
		}
	}
}

// OnWeeklyChange reports the row's POSITION, whatever the row holds. This is
// the assertion that makes "enable Tuesday, save Sunday" unrepresentable.
func TestWeeklyChangeReportsThePosition(t *testing.T) {
	var gotDay int
	var gotRow WeeklyRow
	e := &ScheduleEditor{
		Week:           make([]WeeklyRow, 7),
		OnWeeklyChange: func(d int, r WeeklyRow) { gotDay, gotRow = d, r },
	}
	e.Init(&emptyCtx{})

	e.weeklyChange(2, func(r *WeeklyRow) { r.Active = true })

	if gotDay != 2 {
		t.Errorf("dayOfWeek = %d, want 2 (Tuesday)", gotDay)
	}
	if !gotRow.Active {
		t.Error("the mutation did not reach the reported row")
	}
}

// An inactive day's four time selects are disabled: they are not a schedule,
// they are the absence of one.
func TestInactiveDayDisablesItsTimeSelects(t *testing.T) {
	week := make([]WeeklyRow, 7)
	week[1] = WeeklyRow{Active: true, WorkStart: 480, WorkFinish: 840}
	e := &ScheduleEditor{Week: week}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	// 7 rows x 4 selects = 28; row 1 is active, so 24 are disabled.
	// The attribute serializes as disabled='disabled', so count that marker.
	if got := strings.Count(html, "disabled='disabled'"); got != 24 {
		t.Errorf("disabled selects = %d, want 24:\n%s", got, html)
	}
}

// The header labels every column. Four unlabelled dropdowns was the report.
func TestWeekHeadLabelsEveryColumn(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, string(clsWeekHead)) {
		t.Errorf("the weekly grid has no header row:\n%s", html)
	}
	for _, k := range weekHeadKeys {
		if !strings.Contains(html, k) {
			t.Errorf("header missing column %q:\n%s", k, html)
		}
	}
}
