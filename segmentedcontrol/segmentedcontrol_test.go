//go:build !wasm

package segmentedcontrol

import (
	"strings"
	"testing"
)

func TestSegmentedControl_Render(t *testing.T) {
	comp := &SegmentedControl{
		Options: []Option{
			{Value: "A", Label: "A"},
			{Value: "B", Label: "B"},
		},
		Selected: "A",
	}
	comp.Init(nil)
	html := comp.Render().String()

	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div, got: %s", html)
	}
	if !strings.Contains(html, "role='tablist'") {
		t.Errorf("expected role='tablist', got: %s", html)
	}
	if !strings.Contains(html, "segmentedcontrol") {
		t.Errorf("expected segmentedcontrol class, got: %s", html)
	}
}
