package actionbutton

import (
	"testing"

	"webtyp.com/dom"
)

var _ dom.Component = (*ActionButton)(nil)

func TestActionButton_RenderIdempotent(t *testing.T) {
	b := &ActionButton{Text: "Save", Variant: Primary}

	first := b.Render().String()
	second := b.Render().String()

	if first != second {
		t.Errorf("Render() not idempotent: first=%q second=%q", first, second)
	}
}
