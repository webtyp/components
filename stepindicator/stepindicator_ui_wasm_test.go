//go:build wasm

package stepindicator

import (
	"testing"
	"webtyp.com/dom"
)

func TestStepIndicator_Click(t *testing.T) {
	var clickedIdx int = -1
	comp := &StepIndicator{
		Steps: []Step{
			{Key: "1", Label: "One"},
			{Key: "2", Label: "Two"},
		},
		Active: 0,
		OnChange: func(idx int) {
			clickedIdx = idx
		},
	}
	comp.Init(nil)
	el := comp.Render()

	var btn *dom.Element
	for _, c := range el.ChildNodes() {
		if c.TagName() == "BUTTON" {
			if btn == nil {
				btn = c
			} else {
				btn = c
				break
			}
		}
	}

	if btn != nil {
		btn.DispatchEvent(dom.NewEvent("click"))
	}
}
