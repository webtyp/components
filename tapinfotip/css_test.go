//go:build !wasm

package tapinfotip

import (
	"strings"
	"testing"
)

// The (i) draws at the label's size but must be pressable over 24px.
func TestButtonHasA24pxHitArea(t *testing.T) {
	s := (&TapInfotip{}).RenderCSS().String()
	if !strings.Contains(s, "calc((24px - 1em) / 2)") {
		t.Errorf("button lacks the 24px hit area:\n%s", s)
	}
}
