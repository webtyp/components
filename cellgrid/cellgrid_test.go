//go:build !wasm

package cellgrid

import (
	"strings"
	"testing"
)

func TestCellGrid_ColName(t *testing.T) {
	if ColName(0) != "A" {
		t.Errorf("expected A, got %s", ColName(0))
	}
	if ColName(25) != "Z" {
		t.Errorf("expected Z, got %s", ColName(25))
	}
	if ColName(26) != "AA" {
		t.Errorf("expected AA, got %s", ColName(26))
	}
}

func TestCellGrid_Render(t *testing.T) {
	comp := &CellGrid{
		Cols: 2, Rows: 2,
	}
	comp.Init(nil)
	html := comp.Render().String()

	if !strings.Contains(html, "cellgrid") {
		t.Errorf("expected class 'cellgrid', got: %s", html)
	}
}
