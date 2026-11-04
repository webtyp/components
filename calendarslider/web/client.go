//go:build wasm

package main

import (
	"webtyp.com/components/calendarslider"
	"webtyp.com/components/themetoggle"
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/time"
)

type App struct {
	Element
	cal      calendarslider.CalendarSlider
	selected *SignalString
}

func (a *App) Init(_ Ctx) {
	a.selected = NewString("")
	a.cal = calendarslider.CalendarSlider{
		Zone:      time.MachineZone(), // demo: the visitor's clock; an app passes appkit.BrowserDeps.Zone
		Start:     "2026-08",
		NumMonths: 3,
		Holidays: []calendarslider.Holiday{
			{Date: "2026-08-15", Name: "Asunción de la Virgen"},
			{Date: "2026-09-29", Name: "Victoria de Boquerón"},
		},
		Occupation: []calendarslider.OccupationDay{
			{Date: "2026-08-11", Percent: 60}, {Date: "2026-08-12", Percent: 30}, {Date: "2026-08-13", Percent: 80},
			{Date: "2026-08-17", Percent: 40}, {Date: "2026-08-18", Percent: 100}, {Date: "2026-08-19", Percent: 15},
			{Date: "2026-08-20", Percent: 55}, {Date: "2026-08-21", Percent: 90},
			{Date: "2026-09-01", Percent: 25}, {Date: "2026-09-02", Percent: 70}, {Date: "2026-09-03", Percent: 45},
		},
		OnSelect: func(date string) { a.selected.Set(date) },
	}
}

func (a *App) Render() *Element {
	sel := Span().ID("app-result").BindTextFunc(func() string {
		if d := a.selected.Get(); d != "" {
			return "Día seleccionado: " + d
		}
		return "Día seleccionado: —"
	})

	return Div().Child(
		H1().Text("CalendarSlider — demo"),
		P().Text("Empieza en agosto 2026. Desliza con ‹ › (o arrastra) para ver los meses siguientes. Los días con porcentaje son seleccionables; feriados y domingos en rojo."),
		&a.cal,
		sel,
	)
}

func main() {
	ts := &themetoggle.ThemeToggle{}
	Render("app", &App{})
	Append("body", ts)
	select {}
}
