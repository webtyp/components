//go:build wasm

package cellgrid

import (
	"testing"
	"webtyp.com/dom"
)

func TestCellGrid_Click(t *testing.T) {
	var clickedR, clickedC int = -1, -1
	comp := &CellGrid{
		Cols: 2, Rows: 2,
		OnCellClick: func(r, c int) {
			clickedR = r
			clickedC = c
		},
	}
	comp.Init(nil)
	el := comp.Render()

	// Finding the first cell
	var cell *dom.Element
	for _, c := range el.ChildNodes() {
		if c.TagName() == "DIV" { // simplified child logic
			cell = c
			break
		}
	}

	if cell != nil {
		cell.DispatchEvent(dom.NewEvent("click"))
	}
}
