package chipgrid

import (
	"webtyp.com/dom"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameChipGrid is the widget name for chipgrid.
const NameChipGrid = widget.Name("chipgrid")

const (
	PartChip  = widget.Part("chip")
	PartEmpty = widget.Part("empty")
)

// Chip is one choice.
type Chip struct {
	Value    string
	Label    string
	Disabled bool
}

// ChipGrid shows its chips in a grid and lets the user pick one.
type ChipGrid struct {
	dom.Element
	Selected *dom.SignalString
	Empty    lang.Text
	OnPick   func(value string)
}

func (g *ChipGrid) WidgetName() widget.Name { return NameChipGrid }
func (g *ChipGrid) WidgetKind() widget.Kind { return widget.Listbox }

func (g *ChipGrid) Init(_ dom.Ctx)        { panic("not implemented") }
func (g *ChipGrid) Render() *dom.Element  { panic("not implemented") }
func (g *ChipGrid) SetChips(chips []Chip) { panic("not implemented") }
