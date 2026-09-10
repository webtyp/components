//go:build !wasm

package searchbar

import (
	"strings"
	"testing"
)

func TestSearchBar_RenderCSSEmitsEveryPart(t *testing.T) {
	css := (&SearchBar{}).RenderCSS().String()

	for _, want := range []string{".searchbar ", ".searchbar__icon", ".searchbar__input"} {
		if !strings.Contains(css, want) {
			t.Errorf("stylesheet missing %q\n%s", want, css)
		}
	}

	// The cap sizes its own glyph, so there is no glyph PART to emit — the
	// rule hangs off the cap's selector instead. This is the whole point of
	// IconCap: the glyph size stopped being a component-level decision.
	if !strings.Contains(css, ".searchbar__icon > svg") {
		t.Errorf("IconCap should size the magnifier through the cap's child rule\n%s", css)
	}
	if strings.Contains(css, ".searchbar__glyph") {
		t.Errorf("the glyph should not have a part rule of its own any more\n%s", css)
	}

	for _, sel := range []string{".searchbar ", ".searchbar__input"} {
		if !cssBlockSetsControlHeight(css, sel) {
			t.Errorf("%s must contain a block that sets min-height: var(--control-height\n%s", sel, css)
		}
	}
}

func cssBlockSetsControlHeight(css, sel string) bool {
	for _, block := range cssBlocks(css, sel) {
		if strings.Contains(block, "min-height: var(--control-height") {
			return true
		}
	}
	return false
}

func cssBlocks(css, sel string) []string {
	sel = strings.TrimRight(sel, " ")
	needle := sel + " {"
	var blocks []string
	from := 0
	for {
		i := strings.Index(css[from:], needle)
		if i < 0 {
			break
		}
		i += from
		j := strings.Index(css[i:], "}")
		if j < 0 {
			break
		}
		blocks = append(blocks, css[i:i+j])
		from = i + j
	}
	return blocks
}
