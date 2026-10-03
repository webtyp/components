//go:build !wasm

package stepindicator

import (
	"strings"
	"testing"
)

func TestStepIndicator_Render(t *testing.T) {
	comp := &StepIndicator{
		Steps: []Step{
			{Key: "1", Label: "One"},
			{Key: "2", Label: "Two"},
		},
		Active: 0,
	}
	comp.Init(nil)

	html := comp.Render().String()

	if !strings.HasPrefix(html, "<nav") {
		t.Errorf("expected <nav>, got: %s", html)
	}

	if !strings.Contains(html, "stepindicator") {
		t.Errorf("expected class 'stepindicator', got: %s", html)
	}
}
