//go:build wasm

package calendarslider

import (
	"testing"

	"syscall/js"
	. "webtyp.com/dom"
	"webtyp.com/time"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func domDoc() js.Value { return js.Global().Get("document") }

func query(t *testing.T, sel string) js.Value {
	t.Helper()
	el := domDoc().Call("querySelector", sel)
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("no se encontró %q", sel)
	}
	return el
}

func exists(sel string) bool {
	el := domDoc().Call("querySelector", sel)
	return !el.IsNull() && !el.IsUndefined()
}

// TestDayClickSelects cubre el uso principal: clic en un día ocupable lo
// selecciona (señal + callback) y un día no ocupable no.
func TestDayClickSelects(t *testing.T) {
	var got string
	var gotFilter string
	c := &CalendarSlider{
		Start:      "2026-08",
		Holidays:   []Holiday{{Date: "2026-08-15", Name: "Asunción"}},
		Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 60}},
		OnSelect:   func(date string) { got = date },
	}
	c.OnFilterChange(func(term string) { gotFilter = term })
	c.Init(nil)
	Render("app", c.Render())

	bookable := query(t, "[data-date='2026-08-11'] .calendarslider__day-button")
	bookable.Call("click")
	if c.Selected.Get() != "2026-08-11" {
		t.Errorf("Selected = %q, want 2026-08-11", c.Selected.Get())
	}
	if got != "2026-08-11" {
		t.Errorf("OnSelect = %q, want 2026-08-11", got)
	}
	if gotFilter != "2026-08-11" {
		t.Errorf("OnFilterChange = %q, want 2026-08-11", gotFilter)
	}
	if query(t, "[data-date='2026-08-11']").Call("getAttribute", "data-selected").String() != "true" {
		t.Error("el día seleccionado debería llevar data-selected=true")
	}

	// Un sábado sin ocupación no es seleccionable: ni señal ni callback.
	c.Selected.Set("")
	got = ""
	gotFilter = ""
	// Un día no ocupable ni siquiera tiene botón: no hay nada que pulsar.
	if exists("[data-date='2026-08-01'] .calendarslider__day-button") {
		t.Error("un sábado sin ocupación no debería renderizar botón")
	}
	query(t, "[data-date='2026-08-01']").Call("click")
	if c.Selected.Get() != "" || got != "" || gotFilter != "" {
		t.Error("un día no ocupable no debería seleccionarse")
	}
}

// TestAllMonthsAlwaysInDOM cubre el cambio de arquitectura frente al viejo
// deslizador infinito: los NumMonths meses son hijos estáticos, ninguno se
// desmonta ni se recicla — el slide es puramente visual (scroll-snap).
func TestAllMonthsAlwaysInDOM(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	Render("app", c.Render())

	for _, id := range []string{"[data-month='2026-08']", "[data-month='2026-09']", "[data-month='2026-10']"} {
		if !exists(id) {
			t.Errorf("%s debería existir en el DOM", id)
		}
	}
	if label := query(t, "[data-month='2026-08'] .calendarslider__month-name").Get("textContent").String(); label != "August 2026" {
		t.Fatalf("etiqueta de agosto = %q, want August 2026", label)
	}
	if label := query(t, "[data-month='2026-10'] .calendarslider__month-name").Get("textContent").String(); label != "October 2026" {
		t.Fatalf("etiqueta de octubre = %q, want October 2026", label)
	}
}

// TestNavButtonsSlideToNeighbor cubre la navegación: ‹ › son <button>s; su
// handler llama a slideToMonth, que salta el scroll-snap al mes vecino sin
// tocar location.hash — un <a href="#cs-m-..."> lo mutaría, y un shell con
// enrutado por hash (platformd) lo leería como cambio de ruta, no como
// scroll.
func TestNavButtonsSlideToNeighbor(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	Render("app", c.Render())

	for _, sel := range []string{
		"[data-month='2026-08'] .calendarslider__prev",
		"[data-month='2026-09'] .calendarslider__prev",
		"[data-month='2026-10'] .calendarslider__prev",
		"[data-month='2026-08'] .calendarslider__next",
		"[data-month='2026-09'] .calendarslider__next",
		"[data-month='2026-10'] .calendarslider__next",
	} {
		if tag := query(t, sel).Get("tagName").String(); tag != "BUTTON" {
			t.Errorf("%s debería ser un <button>, tagName = %q", sel, tag)
		}
	}

	// Bucle infinito: agosto (primero) y octubre (último) también tienen
	// botón hacia el otro extremo — nunca hay que recorrer los N meses en
	// orden para volver al principio.
	if !exists("[data-month='2026-08'] .calendarslider__prev") {
		t.Error("el primer mes debería tener botón 'prev' (envuelve al último)")
	}
	if !exists("[data-month='2026-10'] .calendarslider__next") {
		t.Error("el último mes debería tener botón 'next' (envuelve al primero)")
	}

	// El hash no debe cambiar al navegar: el botón vive dentro del widget y
	// el slide es ScrollIntoView, nunca un salto de ancla.
	before := js.Global().Get("location").Get("hash").String()
	query(t, "[data-month='2026-08'] .calendarslider__next").Call("click")
	if after := js.Global().Get("location").Get("hash").String(); after != before {
		t.Errorf("navegar con ‹ › no debería tocar location.hash: antes %q, después %q", before, after)
	}

	// Ningún mes se desmontó durante la navegación.
	for _, id := range []string{"[data-month='2026-08']", "[data-month='2026-09']", "[data-month='2026-10']"} {
		if !exists(id) {
			t.Errorf("%s debería seguir existiendo tras navegar", id)
		}
	}
}

// TestExternalSelectedHighlights cubre el enlace inverso: el host escribe la
// señal y el DOM se pinta sin tocar el calendario.
func TestExternalSelectedHighlights(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08", Occupation: []OccupationDay{{Date: "2026-08-11", Percent: 60}}}
	c.Init(nil)
	Render("app", c.Render())

	c.Selected.Set("2026-08-11")
	if query(t, "[data-date='2026-08-11']").Call("getAttribute", "data-selected").String() != "true" {
		t.Error("escribir Selected debería pintar data-selected en el DOM")
	}

	// Al navegar (el botón ‹ › llama a slideToMonth, que salta el snap sin
	// tocar location.hash), la selección externa sobrevive.
	before := js.Global().Get("location").Get("hash").String()
	query(t, ".calendarslider__next").Call("click")
	if after := js.Global().Get("location").Get("hash").String(); after != before {
		t.Errorf("navegar con ‹ › no debería tocar location.hash: antes %q, después %q", before, after)
	}
	if query(t, "[data-date='2026-08-11']").Call("getAttribute", "data-selected").String() != "true" {
		t.Error("la selección debería sobrevivir a la navegación")
	}
}

// TestWrapNavigationJumpsInstantly cubre el bug real: el ‹ del primer mes
// (envuelve al último) y el › del último (envuelve al primero) deben saltar
// sin animación — un scroll suave ahí viaja visualmente en la dirección
// contraria a través de todos los meses intermedios, lo que se lee como un
// reinicio. La navegación entre meses vecinos sigue siendo suave.
func TestWrapNavigationJumpsInstantly(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08", NumMonths: 3}
	c.Init(nil)
	Render("app", c.Render())

	var lastBehavior string
	proto := js.Global().Get("Element").Get("prototype")
	original := proto.Get("scrollIntoView")
	spy := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			lastBehavior = args[0].Get("behavior").String()
		}
		return nil
	})
	proto.Set("scrollIntoView", spy)
	defer func() {
		proto.Set("scrollIntoView", original)
		spy.Release()
	}()

	query(t, "[data-month='2026-08'] .calendarslider__prev").Call("click")
	if lastBehavior != "instant" {
		t.Errorf("wrap prev (agosto -> octubre) behavior = %q, want instant", lastBehavior)
	}

	query(t, "[data-month='2026-08'] .calendarslider__next").Call("click")
	if lastBehavior != "smooth" {
		t.Errorf("adjacent next (agosto -> septiembre) behavior = %q, want smooth", lastBehavior)
	}

	query(t, "[data-month='2026-10'] .calendarslider__next").Call("click")
	if lastBehavior != "instant" {
		t.Errorf("wrap next (octubre -> agosto) behavior = %q, want instant", lastBehavior)
	}
}

// TestCollapsedChipTogglesEverywhere cubre el chip como toggle de plegado
// en ambos modos: visible también expandido (es el control de plegado, no
// solo el estado colapsado); un tap pliega, otro despliega. A nivel de
// atributo data-current de la tira — sin leer viewport, porque el mecanismo
// es idéntico en desktop y mobile.
func TestCollapsedChipTogglesEverywhere(t *testing.T) {
	c := &CalendarSlider{Start: "2026-08"}
	c.Init(nil)
	Render("app", c.Render())

	stripState := func() string {
		v := query(t, ".calendarslider__strip").Call("getAttribute", "data-current")
		if v.IsNull() || v.IsUndefined() {
			return ""
		}
		return v.String()
	}
	if got := stripState(); got != "true" {
		t.Fatalf("recién montado (expandido) data-current = %q, want true", got)
	}
	query(t, ".calendarslider__collapsed").Call("click")
	if got := stripState(); got == "true" {
		t.Errorf("tras plegar con el chip data-current = %q, want ausente", got)
	}
	query(t, ".calendarslider__collapsed").Call("click")
	if got := stripState(); got != "true" {
		t.Errorf("tras desplegar con el chip data-current = %q, want true", got)
	}
}

// TestChipCollapseDefaultsToToday cubre el plegado con el chip sin día
// elegido: plegar así equivale a tocar hoy — fija Selected a hoy (nunca una
// casilla en blanco) con sus callbacks, y la tira se oculta igual.
func TestChipCollapseDefaultsToToday(t *testing.T) {
	var gotSelect, gotFilter string
	c := &CalendarSlider{Start: "2026-08"}
	c.OnSelect = func(date string) { gotSelect = date }
	c.OnFilterChange(func(term string) { gotFilter = term })
	c.Init(nil)
	Render("app", c.Render())

	today := time.FormatDate(time.Now())
	query(t, ".calendarslider__collapsed").Call("click")
	if c.Selected.Get() != today {
		t.Errorf("plegar sin selección debería fijar hoy (%s), Selected = %q", today, c.Selected.Get())
	}
	if gotSelect != today || gotFilter != today {
		t.Errorf("plegar sin selección debería notificar como tocar hoy: OnSelect = %q, filter = %q", gotSelect, gotFilter)
	}
	v := query(t, ".calendarslider__strip").Call("getAttribute", "data-current")
	if !v.IsNull() && v.String() == "true" {
		t.Error("tras plegar la tira debería ocultarse")
	}
}

// TestTwoInstancesShareAPage is the regression this change exists for: app-demo
// mounts two calendars in one render (a reservation filter and the agenda
// editor). Before, both emitted id="cs-m-<month>" and dom.claimID panicked on
// the duplicate. Now neither invents an id, so both mount, and each ‹ ›
// scrolls its OWN strip.
func TestTwoInstancesShareAPage(t *testing.T) {
	cA := &CalendarSlider{Start: "2026-08", NumMonths: 3}
	cB := &CalendarSlider{Start: "2026-08", NumMonths: 3}

	// Mounted as sibling COMPONENTS under one parent — dom gives each its own
	// instance id and Init — exactly how app-demo composes them.
	parent := NewElement("div").Child(cA).Child(cB)
	if err := Render("app", parent); err != nil {
		t.Fatalf("mounting two calendars in one render failed: %v", err)
	}

	// Both August cards are in the DOM — the duplicate-id panic is gone.
	augCards := domDoc().Call("querySelectorAll", "[data-month='2026-08']")
	if n := augCards.Get("length").Int(); n != 2 {
		t.Fatalf("two calendars => two August cards, got %d", n)
	}
	roots := domDoc().Call("querySelectorAll", ".calendarslider")
	if roots.Get("length").Int() != 2 {
		t.Fatalf("expected two calendar roots, got %d", roots.Get("length").Int())
	}
	rootA := roots.Call("item", 0)

	// Clicking instance A's ‹ (August wraps to October) scrolls a node inside
	// A's own subtree, never B's.
	var scrolledMonth string
	var scrolledRootIsA bool
	proto := js.Global().Get("Element").Get("prototype")
	original := proto.Get("scrollIntoView")
	spy := js.FuncOf(func(this js.Value, args []js.Value) any {
		if card := this.Call("closest", "[data-month]"); !card.IsNull() && !card.IsUndefined() {
			scrolledMonth = card.Call("getAttribute", "data-month").String()
		}
		scrolledRootIsA = this.Call("closest", ".calendarslider").Equal(rootA)
		return nil
	})
	proto.Set("scrollIntoView", spy)
	defer func() {
		proto.Set("scrollIntoView", original)
		spy.Release()
	}()

	augCards.Call("item", 0).Call("querySelector", ".calendarslider__prev").Call("click")

	if scrolledMonth != "2026-10" {
		t.Errorf("A's ‹ from August should target October, scrolled %q", scrolledMonth)
	}
	if !scrolledRootIsA {
		t.Error("A's ‹ scrolled a node outside instance A — cross-instance leak")
	}
}
