//go:build wasm

package main

import (
	"webtyp.com/components/themetoggle"
	. "webtyp.com/dom"
	. "webtyp.com/html"
)

type App struct {
	Element
}

func (a *App) Render() *Element {
	return Div().Child(
		H1().Text("ThemeSwitch Demo"),
		P().Text("Use the control below to cycle through themes."),
		&themetoggle.ThemeToggle{},
	)
}

func main() {
	// ThemeToggle is initialized internally
	Render("app", &App{})
	select {}
}
