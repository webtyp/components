package chipgrid

import (
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameChipGrid is the widget name for chipgrid.
const NameChipGrid = widget.Name("chipgrid")

const (
	PartChip  = widget.Part("chip")
	PartEmpty = widget.Part("empty") // shown when there are no chips
)

var (
	clsRoot  = NameChipGrid.Root()
	clsChip  = NameChipGrid.Class(PartChip)
	clsEmpty = NameChipGrid.Class(PartEmpty)
)

// textNoOptions is the empty text when the host sets none.
const textNoOptions lang.Text = "No options available"

// Chip is one choice. Label is shown as is (it is data, e.g. "09:30", not UI text).
type Chip struct {
	Value    string
	Label    string
	Disabled bool
}

// ChipGrid shows its chips in a grid and lets the user pick one. It is a
// WAI-ARIA listbox: one tab stop (the selected chip, else the first enabled
// one), arrows and Home/End move between enabled chips, Enter/Space pick.
type ChipGrid struct {
	dom.Element // value embed — NEVER pointer (TinyGo heap constraint)

	// Selected holds the picked value, or "". Optional: created if nil so a host
	// can share it.
	Selected *dom.SignalString
	// Empty is the text shown when there are no chips. English; translated in
	// Render. Zero value: "No options available".
	Empty lang.Text
	// OnPick runs when the user picks an enabled chip. Optional.
	OnPick func(value string)

	nodes   *dom.SignalNodes
	chips   []chipRef         // the current chips in DOM order; a slice, never a map
	tabStop *dom.SignalString // value of the chip with tabindex 0 (roving tabindex)
}

// chipRef is one rendered chip.
type chipRef struct {
	chip Chip
	el   *dom.Element
}

func (g *ChipGrid) WidgetName() widget.Name { return NameChipGrid }
func (g *ChipGrid) WidgetKind() widget.Kind { return widget.Listbox }

// Init creates the signals. Calling it again is a no-op.
func (g *ChipGrid) Init(_ dom.Ctx) {
	if g.nodes != nil {
		return
	}
	if g.Selected == nil {
		g.Selected = dom.NewString("")
	}
	g.nodes = dom.NewNodes()
	g.tabStop = dom.NewString("")
	g.show(nil) // the initial selection is the host's: only SetChips may clear it
}

func (g *ChipGrid) Render() *dom.Element {
	g.Init(nil)
	return html.Div().Set(clsRoot.AsAttr()).
		Attr("role", "listbox").
		BindChildren(g.nodes)
}

// SetChips replaces the choices. If Selected is no longer among the enabled
// chips it is cleared to "".
func (g *ChipGrid) SetChips(chips []Chip) {
	g.Init(nil)
	if !g.isEnabled(chips, g.Selected.Get()) {
		g.Selected.Set("")
	}
	g.show(chips)
}

// show renders chips (or the empty text) with the tab stop on the selected
// chip, else the first enabled one.
func (g *ChipGrid) show(chips []Chip) {
	if len(chips) == 0 {
		g.chips = nil
		g.tabStop.Set("")
		empty := g.Empty
		if empty == "" {
			empty = textNoOptions
		}
		g.nodes.Set([]*dom.Element{html.P().Set(clsEmpty.AsAttr()).Text(lang.Translate(empty).String())})
		return
	}

	g.chips = make([]chipRef, len(chips))
	stop := ""
	for _, c := range chips {
		if !c.Disabled && (c.Value == g.Selected.Get() || stop == "") {
			stop = c.Value
		}
	}
	g.tabStop.Set(stop)
	nodes := make([]*dom.Element, len(chips))
	for i, c := range chips {
		g.chips[i] = chipRef{chip: c, el: g.buildChip(i, c)}
		nodes[i] = g.chips[i].el
	}
	g.nodes.Set(nodes)
}

func (g *ChipGrid) isEnabled(chips []Chip, value string) bool {
	for _, c := range chips {
		if c.Value == value && !c.Disabled {
			return true
		}
	}
	return false
}

func (g *ChipGrid) buildChip(i int, c Chip) *dom.Element {
	value := c.Value
	selected := dom.DeriveBool(func() bool { return g.Selected.Get() == value })
	ariaSelected := dom.DeriveString(func() string {
		if selected.Get() {
			return "true"
		}
		return "false"
	})
	tab := dom.DeriveString(func() string {
		if g.tabStop.Get() == value {
			return "0"
		}
		return "-1"
	})
	el := html.Button().Set(clsChip.AsAttr()).
		Key(value).
		Attr("type", "button").
		Attr("role", "option").
		Attr("data-value", value).
		BindAttr("tabindex", tab).
		BindAttr("aria-selected", ariaSelected).
		BindState(widget.Selected, selected).
		BindState(widget.Disabled, dom.NewBool(c.Disabled)).
		Text(c.Label)
	if c.Disabled {
		el.Attr("disabled", "")
	}
	el.OnClick(func(dom.Event) { g.pick(i) })
	el.OnKeyDown(func(e dom.KeyEvent) { g.onKey(i, e) })
	return el
}

// pick selects chip i if it is enabled and moves the tab stop to it.
func (g *ChipGrid) pick(i int) {
	c := g.chips[i].chip
	if c.Disabled {
		return
	}
	g.Selected.Set(c.Value)
	g.moveTabStop(i, false)
	if g.OnPick != nil {
		g.OnPick(c.Value)
	}
}

func (g *ChipGrid) onKey(i int, e dom.KeyEvent) {
	next := -1
	switch e.Key() {
	case dom.KeyArrowRight, dom.KeyArrowDown:
		next = g.enabledFrom(i+1, 1)
	case dom.KeyArrowLeft, dom.KeyArrowUp:
		next = g.enabledFrom(i-1, -1)
	case dom.KeyHome:
		next = g.enabledFrom(0, 1)
	case dom.KeyEnd:
		next = g.enabledFrom(len(g.chips)-1, -1)
	case dom.KeyEnter, dom.KeySpace:
		e.PreventDefault()
		g.pick(i)
		return
	default:
		return
	}
	if next < 0 {
		return
	}
	e.PreventDefault()
	g.moveTabStop(next, true)
}

// enabledFrom returns the first enabled chip from start walking by step, or -1.
func (g *ChipGrid) enabledFrom(start, step int) int {
	for i := start; i >= 0 && i < len(g.chips); i += step {
		if !g.chips[i].chip.Disabled {
			return i
		}
	}
	return -1
}

// moveTabStop moves the roving tabindex to chip i, and the DOM focus when focus is true.
func (g *ChipGrid) moveTabStop(i int, focus bool) {
	g.tabStop.Set(g.chips[i].chip.Value)
	if !focus {
		return
	}
	if ref, ok := g.chips[i].el.Ref(); ok {
		ref.Focus()
	}
}
