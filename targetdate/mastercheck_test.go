//go:build !wasm

package targetdate

import (
	"strings"
	"testing"
)

// classRules concatenates every rule BODY in css whose selector list mentions
// class (a bare part name like ".targetdate__sel-count"). The generator
// groups parts that share identical declarations into one selector list and
// spreads a single part's OWN declarations across several such rules, so a
// part's full declaration set is scattered across multiple "{...}" blocks —
// a single strings.Index(css, class+" {") only ever finds the first one.
func classRules(css, class string) string {
	var out strings.Builder
	for i := 0; ; {
		j := strings.Index(css[i:], class)
		if j == -1 {
			break
		}
		start := i + j
		open := strings.Index(css[start:], "{")
		end := strings.Index(css[start:], "}")
		if open == -1 || end == -1 {
			break
		}
		out.WriteString(css[start+open : start+end])
		out.WriteString("\n")
		i = start + end + 1
	}
	return out.String()
}

// The selection header is the widget root's first child. Unlike its
// content, the strip itself is NEVER display:none, so the list does not
// shift when selection mode opens or closes. It is the strip
// listselect.Header builds; the widget merely Child()s it above the <ul>.
// Its reserved height is not a bespoke min-height: sel-spacer is the one
// child that stays visible in every mode, sized like the select-all icon,
// so the strip's own flex height follows that icon's footprint.
func TestTargetDate_MasterCheckHeaderReservesHeightAlways(t *testing.T) {
	css := (&TargetDate{}).RenderCSS().String()
	header := classRules(css, ".targetdate__sel-header")
	if header == "" {
		t.Fatal("expected a rule for .targetdate__sel-header")
	}
	if strings.Contains(header, "display: none") {
		t.Errorf("the header strip itself must never be display:none, block:\n%s", header)
	}
	spacer := classRules(css, ".targetdate__sel-spacer")
	if spacer == "" {
		t.Fatal("expected a rule for .targetdate__sel-spacer")
	}
	if strings.Contains(spacer, "display: none") {
		t.Errorf("the spacer must stay visible in every mode — it's what reserves the strip's height, block:\n%s", spacer)
	}
	if !strings.Contains(spacer, "width:") || !strings.Contains(spacer, "height:") {
		t.Errorf("the spacer must carry its own icon-sized footprint, block:\n%s", spacer)
	}
}

// The select-all box, unlike the count, IS hidden until selection mode
// opens: it is an action control that means nothing without selection mode.
func TestTargetDate_MasterCheckBoxHiddenUntilSelectionMode(t *testing.T) {
	css := (&TargetDate{}).RenderCSS().String()
	body := classRules(css, ".targetdate__sel-all")
	if body == "" {
		t.Fatal("expected a rule for .targetdate__sel-all")
	}
	if !strings.Contains(body, "display: none") {
		t.Errorf("sel-all must be hidden by default, block:\n%s", body)
	}
	if !strings.Contains(css, `.targetdate[data-open="true"] .targetdate__sel-all {`) {
		t.Errorf("sel-all must be revealed by the list's open state")
	}
}

// The count, unlike the box, is NEVER hidden: "k / N" is useful outside
// selection mode too (N alone answers "how many records are there"), so it
// shows even before anything is marked.
func TestTargetDate_MasterCheckCountAlwaysVisible(t *testing.T) {
	css := (&TargetDate{}).RenderCSS().String()
	body := classRules(css, ".targetdate__sel-count")
	if body == "" {
		t.Fatal("expected a rule for .targetdate__sel-count")
	}
	if strings.Contains(body, "display: none") {
		t.Errorf("sel-count must never be display:none, block:\n%s", body)
	}
	if !strings.Contains(body, "display: flex") || !strings.Contains(body, "justify-content: center") {
		t.Errorf("sel-count must be a centred flex box outside any state rule, block:\n%s", body)
	}
}

// Tapping the select-all box with nothing / some marked selects every row;
// tapping it with all marked clears. The click handler lives in
// listselect.Header (covered by its WASM test); here only the widget's wiring
// is checked — the Mode the header drives.
func TestTargetDate_MasterCheckTogglesAll(t *testing.T) {
	td := &TargetDate{}
	td.Init(nil)
	td.SetItems([]Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})
	td.SetSelectMode(true)

	if n := td.sel.Count(); n > 0 && n == 3 {
		td.sel.Clear()
	} else {
		td.sel.CheckAll(td.itemIDs())
	}
	if td.sel.Count() != 3 {
		t.Fatalf("first tap must select all, Count = %d", td.sel.Count())
	}
	if n := td.sel.Count(); n > 0 && n == 3 {
		td.sel.Clear()
	} else {
		td.sel.CheckAll(td.itemIDs())
	}
	if td.sel.Count() != 0 {
		t.Fatalf("second tap must clear, Count = %d", td.sel.Count())
	}
}

// The count label renders "n / total" inside the header strip.
func TestTargetDate_MasterCheckShowsNOfTotal(t *testing.T) {
	td := &TargetDate{}
	td.Init(nil)
	td.SetItems([]Item{{ID: "1"}, {ID: "2"}})
	td.SetSelectMode(true)
	td.sel.CheckAll([]string{"1"})

	html := td.Render().String()
	if !strings.Contains(html, "1 / 2") {
		t.Errorf("master check must show \"1 / 2\", got:\n%s", html)
	}
}

// The header reuses the shared glyphs, not a bespoke tick.
func TestTargetDate_MasterCheckUsesSharedGlyphs(t *testing.T) {
	td := &TargetDate{}
	td.Init(nil)
	html := td.Render().String()
	// The box carries one fixed glyph — never trash/pencil, which name the
	// ACTION a footer commit button already shows; select-all names the
	// SELECTION and stays put regardless of what it's used for.
	if !strings.Contains(html, `href='#selectall'`) {
		t.Errorf("master check must reference the shared selectall glyph:\n%s", html)
	}
}

// A <use href='#selectall'> with no matching <symbol> renders nothing — the
// exact defect a missing IconSvg() registration produces. This is the test
// TestTargetDate_MasterCheckUsesSharedGlyphs cannot catch: it only proves the
// reference exists, not that the symbol backing it does.
func TestTargetDate_IconSvgRegistersSelectAll(t *testing.T) {
	out := (&TargetDate{}).IconSvg().String()
	if !strings.Contains(out, `<symbol id="selectall"`) {
		t.Errorf("IconSvg must register the selectall symbol the header references:\n%s", out)
	}
}
