package stepindicator

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/widget"
)

// NameStepIndicator is the widget name for stepindicator.
const NameStepIndicator = widget.Name("stepindicator")

const (
	PartStep  = widget.Part("step")
	PartBadge = widget.Part("badge")
	PartLabel = widget.Part("label")
	PartLine  = widget.Part("line")
)

var (
	clsNav   = NameStepIndicator.Root()
	clsStep  = NameStepIndicator.Class(PartStep)
	clsBadge = NameStepIndicator.Class(PartBadge)
	clsLabel = NameStepIndicator.Class(PartLabel)
	clsLine  = NameStepIndicator.Class(PartLine)
)

type Step struct {
	Key   string
	Label string
}

type StepIndicator struct {
	dom.Element
	Steps    []Step
	Active   int // 0-based active step index
	OnChange func(index int)

	activeSig *dom.SignalString
}

func (c *StepIndicator) WidgetName() widget.Name { return NameStepIndicator }
func (c *StepIndicator) WidgetKind() widget.Kind { return widget.Tabs }

func (c *StepIndicator) Init(_ dom.Ctx) {
	c.activeSig = dom.NewString(fmt.Sprint(c.Active))
}

func (c *StepIndicator) Render() *dom.Element {
	if c.activeSig == nil {
		c.Init(nil)
	}

	nav := html.Nav().Set(clsNav.AsAttr()).Attr("role", "navigation")

	for i, step := range c.Steps {
		idx := i
		idxStr := fmt.Sprint(idx)

		badge := html.Div().Set(clsBadge.AsAttr()).
			BindTextFunc(func() string {
				cur, _ := parse(c.activeSig.Get())
				if cur > idx {
					return "✓"
				}
				return fmt.Sprint(idx + 1)
			})

		label := html.Span().Set(clsLabel.AsAttr()).Text(step.Label)

		btn := html.Button().Set(clsStep.AsAttr()).
			BindStateFunc(widget.Current, func() bool { return c.activeSig.Get() == idxStr }).
			BindStateFunc(widget.Selected, func() bool { cur, _ := parse(c.activeSig.Get()); return cur > idx }).
			Attr("type", "button").
			Child(badge).
			Child(label).
			OnClick(func(dom.Event) {
				c.activeSig.Set(idxStr)
				if c.OnChange != nil {
					c.OnChange(idx)
				}
			})

		nav.Child(btn)

		if i < len(c.Steps)-1 {
			line := html.Div().Set(clsLine.AsAttr()).
				BindStateFunc(widget.Current, func() bool { cur, _ := parse(c.activeSig.Get()); return cur > idx })
			nav.Child(line)
		}
	}

	return nav
}

func parse(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	res := 0
	for i := 0; i < len(s); i++ {
		res = res*10 + int(s[i]-'0')
	}
	return res, nil
}
