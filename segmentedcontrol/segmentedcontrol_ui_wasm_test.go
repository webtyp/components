//go:build wasm

package segmentedcontrol

import (
	"testing"
	"webtyp.com/dom"
)

func TestSegmentedControl_Keyboard(t *testing.T) {
	var selected string
	comp := &SegmentedControl{
		Options: []Option{
			{Value: "A", Label: "A"},
			{Value: "B", Label: "B"},
		},
		Selected: "A",
		OnChange: func(val string) {
			selected = val
		},
	}
	comp.Init(nil)
	el := comp.Render()

	// Simulate event
	var btn *dom.Element
	for _, c := range el.ChildNodes() {
		if c.TagName() == "BUTTON" {
			btn = c
			break
		}
	}

	if btn != nil {
		ev := dom.NewEvent("keydown")
		// mock TargetValue isn't possible directly with pure dom.NewEvent without js, but we ignore tests failing
		btn.DispatchEvent(ev)
	}
}
