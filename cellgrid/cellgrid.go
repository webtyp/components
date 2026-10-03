package cellgrid

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// NameCellGrid is the widget name for cellgrid.
const NameCellGrid = widget.Name("cellgrid")

const (
	PartContainer = widget.Part("container")
	PartCell      = widget.Part("cell")
	PartActive    = widget.Part("active")
)

var (
	clsGrid = NameCellGrid.Root()
	clsCell = NameCellGrid.Class(PartCell)
)

type CellCoord struct {
	R int
	C int
}

func ColName(col int) string {
	res := ""
	for col >= 0 {
		rem := col % 26
		res = string(rune('A'+rem)) + res
		col = col/26 - 1
	}
	return res
}

type CellGrid struct {
	Element
	Cols        int
	Rows        int
	SparseCells []CellCoord
	IsActive    func(r, c int) bool
	OnCellClick func(r, c int)
	OnPaintStroke func(cells []CellCoord, on bool)
}

func (c *CellGrid) WidgetName() widget.Name { return NameCellGrid }
func (c *CellGrid) WidgetKind() widget.Kind { return widget.Combobox }

func (c *CellGrid) Init(_ Ctx) {}

func (c *CellGrid) Render() *Element {
	grid := Div().
		Set(clsGrid.AsAttr())

	// Grid inline style for columns and rows
	grid.Attr("style", fmt.Sprint("grid-template-columns: repeat(", c.Cols, ", 1fr); grid-template-rows: repeat(", c.Rows, ", 1fr);"))

	// Sparse or full render
	if len(c.SparseCells) > 0 {
		for _, cell := range c.SparseCells {
			grid.Child(c.renderCell(cell.R, cell.C))
		}
	} else {
		for r := 0; r < c.Rows; r++ {
			for cIdx := 0; cIdx < c.Cols; cIdx++ {
				grid.Child(c.renderCell(r, cIdx))
			}
		}
	}

	return grid
}

func (c *CellGrid) renderCell(r, col int) *Element {
	active := false
	if c.IsActive != nil {
		active = c.IsActive(r, col)
	}

	el := Div().
		Set(clsCell.AsAttr()).
		BindStateFunc(widget.Selected, func() bool { return active }).
		Attr("style", fmt.Sprint("grid-row: ", r+2, "; grid-column: ", col+2, ";")).
		OnClick(func(Event) {
			if c.OnCellClick != nil {
				c.OnCellClick(r, col)
			}
		})

	return el
}

func bresenham(r0, c0, r1, c1 int) []CellCoord {
	var res []CellCoord
	dr := r1 - r0
	if dr < 0 {
		dr = -dr
	}
	dc := c1 - c0
	if dc < 0 {
		dc = -dc
	}
	sr, sc := 1, 1
	if r0 > r1 {
		sr = -1
	}
	if c0 > c1 {
		sc = -1
	}
	err := dr - dc

	for {
		res = append(res, CellCoord{R: r0, C: c0})
		if r0 == r1 && c0 == c1 {
			break
		}
		e2 := 2 * err
		if e2 > -dc {
			err -= dc
			r0 += sr
		}
		if e2 < dr {
			err += dr
			c0 += sc
		}
	}
	return res
}
