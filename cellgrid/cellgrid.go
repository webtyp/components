package cellgrid

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/widget"
)

// NameCellGrid is the widget name for cellgrid.
const NameCellGrid = widget.Name("cellgrid")

const (
	PartContainer = widget.Part("container")
	PartCell      = widget.Part("cell")
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
	dom.Element
	Cols          int
	Rows          int
	SparseCells   []CellCoord
	IsActive      func(r, c int) bool
	OnCellClick   func(r, c int)
	OnPaintStroke func(cells []CellCoord, on bool)
}

func (c *CellGrid) WidgetName() widget.Name { return NameCellGrid }
func (c *CellGrid) WidgetKind() widget.Kind { return widget.Grid }

func (c *CellGrid) Init(_ dom.Ctx) {}

func (c *CellGrid) Render() *dom.Element {
	grid := html.Div().
		Set(clsGrid.AsAttr())

	grid.Attr("style", fmt.Sprintf("display: grid; grid-template-columns: repeat(%d, 1fr); grid-template-rows: repeat(%d, 1fr);", c.Cols, c.Rows))

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

func (c *CellGrid) renderCell(r, col int) *dom.Element {
	active := false
	if c.IsActive != nil {
		active = c.IsActive(r, col)
	}

	el := html.Div().
		Set(clsCell.AsAttr()).
		BindStateFunc(widget.Selected, func() bool { return active }).
		Attr("style", fmt.Sprintf("grid-row: %d; grid-column: %d;", r+1, col+1)).
		OnClick(func(dom.Event) {
			if c.OnCellClick != nil {
				c.OnCellClick(r, col)
			}
		})

	return el
}

func Bresenham(r0, c0, r1, c1 int) []CellCoord {
	var res []CellCoord
	dr := r1 - r0
	if dr < 0 {
		dr = -dr
	}
	dc := c1 - c0
	if dc < 0 {
		dc = -dc
	}

	sr := 1
	if r0 > r1 {
		sr = -1
	}
	sc := 1
	if c0 > c1 {
		sc = -1
	}

	err := dr - dc
	currR, currC := r0, c0

	for {
		res = append(res, CellCoord{R: currR, C: currC})
		if currR == r1 && currC == c1 {
			break
		}
		e2 := 2 * err
		if e2 > -dc {
			err -= dc
			currR += sr
		}
		if e2 < dr {
			err += dr
			currC += sc
		}
	}

	return res
}
