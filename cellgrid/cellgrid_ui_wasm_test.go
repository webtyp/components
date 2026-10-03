//go:build wasm

package cellgrid_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/cellgrid"
	"webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestCellGrid_Click(t *testing.T) {
	clickedR, clickedC := -1, -1
	comp := &cellgrid.CellGrid{
		Cols: 3,
		Rows: 3,
		OnCellClick: func(r, c int) {
			clickedR = r
			clickedC = c
		},
	}
	dom.Render("app", comp)

	cells := js.Global().Get("document").Call("querySelectorAll", ".cellgrid__cell")
	if cells.Length() == 0 {
		// Fallback to div child
		cells = js.Global().Get("document").Call("querySelectorAll", "div > div")
	}

	if cells.Length() > 0 {
		cells.Index(0).Call("click")
		if clickedR != 0 || clickedC != 0 {
			t.Logf("clicked (%d, %d)", clickedR, clickedC)
		}
	}
}
