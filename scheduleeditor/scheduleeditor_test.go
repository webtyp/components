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
		Week: []WeeklyRow{
			{DayOfWeek: 0, Active: false, WorkStart: 0, WorkFinish: 0},
			{DayOfWeek: 1, Active: true, WorkStart: 540, WorkFinish: 1020, BreakStart: 780, BreakFinish: 840},
			{DayOfWeek: 2, Active: true, WorkStart: 540, WorkFinish: 1020},
			{DayOfWeek: 3, Active: true, WorkStart: 540, WorkFinish: 1020},
			{DayOfWeek: 4, Active: true, WorkStart: 540, WorkFinish: 1020},
			{DayOfWeek: 5, Active: true, WorkStart: 540, WorkFinish: 1020},
			{DayOfWeek: 6, Active: false, WorkStart: 0, WorkFinish: 0},
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
// selects de hora (la fila hours lleva data-open); con HOLIDAY los oculta.
// El form en sí lleva data-open mientras hay día elegido, así que el marcador
// del "horario visible" es que data-open aparezca DOS veces (form + hours).
func TestSpecialHoursShowsTimeSelects(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	e.excType.Set(ExcSpecialHours)
	e.sel.Set("2026-09-19")

	html := e.buildExceptionForm().String()
	if countOpen(html) != 2 {
		t.Errorf("SPECIAL_HOURS must keep the hours row open (2x data-open):\n%s", html)
	}
}

func TestHolidayHidesTimeSelects(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	e.excType.Set(ExcHoliday)
	e.sel.Set("2026-09-19")

	html := e.buildExceptionForm().String()
	if countOpen(html) != 1 {
		t.Errorf("HOLIDAY must hide the hours row (exactly 1 data-open, the form's):\n%s", html)
	}
	if !strings.Contains(html, "scheduleeditor__exc-type") {
		t.Errorf("the type radio group must render:\n%s", html)
	}
}

func countOpen(html string) int {
	n := 0
	for i := 0; i+len("data-open=''") <= len(html); i++ {
		if html[i:i+len("data-open=''")] == "data-open=''" {
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
// y llama OnWeeklyChange con DayOfWeek==1 y los minutos editados.
func TestCallbacks_WeeklyChangeMonday(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})

	var got *WeeklyRow
	e.OnWeeklyChange = func(r WeeklyRow) { got = &r }

	// The weekly select's change handler recomputes from the ORIGINAL row with
	// one field set — exercising the same path the DOM wiring uses. We call the
	// component's action directly (the handler is a small wrapper over it).
	e.weeklyChange(1, func(r *WeeklyRow) { r.WorkStart = 600 })

	if got == nil {
		t.Fatal("OnWeeklyChange not called")
	}
	if got.DayOfWeek != 1 {
		t.Errorf("DayOfWeek = %d, want 1", got.DayOfWeek)
	}
	if got.WorkStart != 600 {
		t.Errorf("WorkStart = %d, want 600", got.WorkStart)
	}
	// Rest of the row preserved.
	if got.WorkFinish != 1020 || got.BreakStart != 780 || got.BreakFinish != 840 {
		t.Errorf("row not preserved: %+v", *got)
	}
}
