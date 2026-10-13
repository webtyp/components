//go:build !wasm

package themetoggle

import (
	"testing"

	"webtyp.com/dom"
)

var _ dom.Component = (*ThemeToggle)(nil)

func TestThemeToggle_InitTwiceSafe(t *testing.T) {
	ts := &ThemeToggle{}

	ts.Init(nil)
	ts.Init(nil)
}

func TestThemeToggle_RenderIdempotent(t *testing.T) {
	ts := &ThemeToggle{}
	ts.Init(nil)

	first := ts.Render().String()
	second := ts.Render().String()

	if first != second {
		t.Errorf("Render() not idempotent: first=%q second=%q", first, second)
	}
}
