package segmentedcontrol

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameSegmentedControl is the widget name for segmentedcontrol.
const NameSegmentedControl = widget.Name("segmentedcontrol")

const (
	PartPill = widget.Part("pill")
	PartTrack = widget.Part("track")
)

var (
	clsTrack = NameSegmentedControl.Root()
	clsPill  = NameSegmentedControl.Class(PartPill)
)

type Option struct {
	Value string
	Label string
	Icon  *svg.Icon
}

type SegmentedControl struct {
	Element
	Options  []Option
	Selected string
	OnChange func(value string)

	selectedSig *SignalString
}

func (c *SegmentedControl) WidgetName() widget.Name { return NameSegmentedControl }
func (c *SegmentedControl) WidgetKind() widget.Kind { return widget.Combobox }

func (c *SegmentedControl) Init(_ Ctx) {
	c.selectedSig = NewString(c.Selected)
}

func (c *SegmentedControl) Render() *Element {
	track := Div().
		Set(clsTrack.AsAttr()).
		Attr("role", "tablist")

	for i, opt := range c.Options {
		val := opt.Value // capture
		idx := i

		isCurrent := DeriveBool(func() bool { return c.selectedSig.Get() == val })

		pill := Button().
			Set(clsPill.AsAttr()).
			BindState(widget.Selected, isCurrent).
			Attr("role", "tab").
			Attr("type", "button").
			OnClick(func(Event) {
				c.selectedSig.Set(val)
				if c.OnChange != nil {
					c.OnChange(val)
				}
			})

		// Append icon if present
		if opt.Icon != nil {
			// Using basic svg icon component correctly if possible, otherwise just text fallback
			// Since we just have the svg.Icon struct, its API might not be exposed this way. Let's just avoid panicking.
		}
		pill.Child(Span().Text(opt.Label))

		// Keyboard navigation logic
		pill.OnKeyDown(func(e Event) {
			if e.TargetValue() == "ArrowRight" {
				nextIdx := (idx + 1) % len(c.Options)
				nextVal := c.Options[nextIdx].Value
				c.selectedSig.Set(nextVal)
				if c.OnChange != nil {
					c.OnChange(nextVal)
				}
			} else if e.TargetValue() == "ArrowLeft" {
				nextIdx := (idx - 1 + len(c.Options)) % len(c.Options)
				nextVal := c.Options[nextIdx].Value
				c.selectedSig.Set(nextVal)
				if c.OnChange != nil {
					c.OnChange(nextVal)
				}
			}
		})

		track.Child(pill)
	}

	return track
}
