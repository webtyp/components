//go:build wasm

package scheduleeditor_test

import (
	"testing"

	"syscall/js"
	"webtyp.com/components/scheduleeditor"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func query(t *testing.T, sel string) js.Value {
	t.Helper()
	el := js.Global().Get("document").Call("querySelector", sel)
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("no se encontró %q", sel)
	}
	return el
}

// TestDayStateChoosesTheAction cubre el cambio de fondo del panel de fechas:
// la acción que se ofrece la decide el ESTADO DEL DÍA, no una lista de tipos.
// Antes el formulario mostraba las cuatro opciones sobre cualquier fecha, y la
// mayoría eran inválidas según el día — "Cerrado" sobre un día que ya está
// cerrado, "Horario especial" sobre una ventana que no existe.
//
// Solo un bloque de acción está abierto a la vez, así que "el botón del bloque
// abierto" es inequívoco por construcción: es lo que este test usa para
// direccionarlos, y es la razón por la que no necesitan una clase cada uno.
func TestDayStateChoosesTheAction(t *testing.T) {
	se := &scheduleeditor.ScheduleEditor{
		// Lunes a viernes en el patrón; sábado y domingo fuera.
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
		Exceptions: []scheduleeditor.Exception{
			{ID: "x1", Date: "2026-09-18", Type: scheduleeditor.ExcHoliday},
		},
		Holidays: []string{},
	}
	se.Init(nil)
	Render("app", se)

	openAction := func() js.Value {
		return js.Global().Get("document").Call("querySelector", ".scheduleeditor__exc-action[data-open='true']")
	}

	// 1. Un día que el patrón cubre y no tiene excepción: se ofrece cerrarlo.
	query(t, "[data-date='2026-09-17'] .calendarslider__day-button").Call("click")
	form := query(t, ".scheduleeditor__exc-form")
	if v := form.Call("getAttribute", "data-open"); v.IsNull() || v.IsUndefined() {
		t.Fatal("elegir un día debería abrir el panel de acción")
	}
	act := openAction()
	if act.IsNull() {
		t.Fatal("un día laborable sin excepción debería ofrecer una acción")
	}
	var added scheduleeditor.Exception
	se.OnExceptionAdd = func(ex scheduleeditor.Exception) { added = ex }
	act.Call("querySelector", ".scheduleeditor__exc-add").Call("click")
	if added.Date != "2026-09-17" || added.Type != scheduleeditor.ExcHoliday {
		t.Fatalf("cerrar un día laborable debería emitir HOLIDAY en su fecha, llegó %+v", added)
	}

	// 2. Un día que YA tiene excepción: la única acción es deshacerla. Antes
	//    volvía a abrir el alta y se podían apilar dos sobre la misma fecha.
	var removed string
	se.OnExceptionRemove = func(id string) { removed = id }
	query(t, "[data-date='2026-09-18'] .calendarslider__day-button").Call("click")
	act = openAction()
	if act.IsNull() {
		t.Fatal("un día con excepción debería ofrecer volver al horario normal")
	}
	act.Call("querySelector", ".scheduleeditor__exc-add").Call("click")
	if removed != "x1" {
		t.Fatalf("volver al horario normal debería quitar la excepción x1, llegó %q", removed)
	}
}

func TestTwoInstancesShareAPage(t *testing.T) {
	se1 := &scheduleeditor.ScheduleEditor{
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
	}
	se2 := &scheduleeditor.ScheduleEditor{
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
	}
	se1.Init(nil)
	se2.Init(nil)

	parent := NewElement("div").Child(se1).Child(se2)
	if err := Render("app", parent); err != nil {
		t.Fatalf("mounting two scheduleeditors in one render failed: %v", err)
	}

	editors := js.Global().Get("document").Call("querySelectorAll", ".scheduleeditor")
	if editors.Get("length").Int() != 2 {
		t.Fatalf("expected 2 scheduleeditors in DOM, got %d", editors.Get("length").Int())
	}
}
