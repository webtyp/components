package segmentedcontrol

import (
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameSegmentedControl is the widget name for segmentedcontrol.
const NameSegmentedControl = widget.Name("segmentedcontrol")

const (
	PartPill  = widget.Part("pill")
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
	dom.Element
	Options  []Option
	Selected string
	OnChange func(value string)

	selectedSig *dom.SignalString
}

func (c *SegmentedControl) WidgetName() widget.Name { return NameSegmentedControl }
func (c *SegmentedControl) WidgetKind() widget.Kind { return widget.Tabs }

func (c *SegmentedControl) Init(_ dom.Ctx) {
	c.selectedSig = dom.NewString(c.Selected)
}

func (c *SegmentedControl) Render() *dom.Element {
	if c.selectedSig == nil {
		c.Init(nil)
	}

	track := html.Div().
		Set(clsTrack.AsAttr()).
		Attr("role", "tablist")

	for i, opt := range c.Options {
		val := opt.Value
		idx := i

		isCurrent := dom.DeriveBool(func() bool { return c.selectedSig.Get() == val })

		pill := html.Button().
			Set(clsPill.AsAttr()).
			BindState(widget.Selected, isCurrent).
			Attr("role", "tab").
			Attr("type", "button").
			OnClick(func(dom.Event) {
				c.selectedSig.Set(val)
				if c.OnChange != nil {
					c.OnChange(val)
				}
			})

		pill.Child(html.Span().Text(opt.Label))

		pill.OnKeyDown(func(e dom.KeyEvent) {
			if e.Key() == dom.KeyArrowRight {
				nextIdx := (idx + 1) % len(c.Options)
				nextVal := c.Options[nextIdx].Value
				c.selectedSig.Set(nextVal)
				if c.OnChange != nil {
					c.OnChange(nextVal)
				}
			} else if e.Key() == dom.KeyArrowLeft {
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
