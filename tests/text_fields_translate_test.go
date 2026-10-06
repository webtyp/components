//go:build wasm

// Root-level test (justified): Public API integration for multiple component packages in WASM.
package components_test

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	"webtyp.com/components/composebar"
	"webtyp.com/components/decktabs"
	"webtyp.com/components/searchbar"
	"webtyp.com/components/selectsearch"
	"webtyp.com/dom"
	"webtyp.com/lang"
)

func TestMain(m *testing.M) {
	// Mount point
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)

	// Dictionary script
	script := js.Global().Get("document").Call("createElement", "script")
	script.Set("type", "application/json")
	script.Set("id", lang.ScriptID)
	// byte-order sorted keys
	script.Set("textContent", `{"default":"es","languages":["es"],"keys":{"Delete":["Eliminar"],"Hours":["Horario"],"Search patients":["Buscar pacientes"],"Send":["Enviar"]}}`)
	js.Global().Get("document").Get("head").Call("appendChild", script)

	m.Run()
}

func TestFixedText_Translated(t *testing.T) {
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)

	// Clean up dom function
	clearApp := func() {
		js.Global().Get("document").Call("getElementById", "app").Set("innerHTML", "")
	}
	defer clearApp()

	// 1. SearchBar (renders synchronously in Render().String())
	html := (&searchbar.SearchBar{Placeholder: "Search patients"}).Render().String()
	if !strings.Contains(html, "placeholder='Buscar pacientes'") {
		t.Errorf("SearchBar: expected placeholder='Buscar pacientes', got: %s", html)
	}

	// 2. ComposeBar
	html = (&composebar.ComposeBar{SendLabel: "Send", OnSend: func(string) {}}).Render().String()
	if !strings.Contains(html, "Enviar") || strings.Contains(html, "Send") {
		t.Errorf("ComposeBar: expected 'Enviar', got: %s", html)
	}

	// 3. DeckTabs
	html = (&decktabs.DeckTabs{Items: []decktabs.Item{{ID: "1", Label: "Hours"}}}).Render().String()
	if !strings.Contains(html, "Horario") || strings.Contains(html, "Hours") {
		t.Errorf("DeckTabs: expected 'Horario', got: %s", html)
	}

	// 4. SelectSearch (requires mounting for Show())
	clearApp()
	ss := &selectsearch.SelectSearch{
		Placeholder: "Search patients",
		Options:     []selectsearch.SsOption{{ID: "1", Label: "Delete"}},
	}
	ss.Init(nil)
	if err := dom.Render("app", ss); err != nil {
		t.Fatalf("failed to render SelectSearch: %v", err)
	}

	// Yield for dom queue and signals
	time.Sleep(50 * time.Millisecond)

	el := js.Global().Get("document").Call("querySelector", "#app .selectsearch__placeholder")
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("SelectSearch placeholder span not found in DOM")
	}

	got := el.Get("textContent").String()
	if got != "Buscar pacientes" {
		t.Errorf("SelectSearch: expected placeholder 'Buscar pacientes', got: %s", got)
	}
}
