//go:build !wasm

package targetdate

import (
	"strings"
	"testing"
	"webtyp.com/time"
)

// Mirror of targetlist's danger contract: the two lists stay
// interchangeable for crudview, and a state that exists in one and not the
// other is exactly how they drift apart.
func TestTargetDate_DangerMarksInvalidNotSelected(t *testing.T) {
	td := &TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	row := td.buildRow(Item{ID: "1", Label: "Row 1"})

	td.SetSelectMode(true)
	td.SetDanger(true)
	td.sel.Toggle("1")

	html := row.String()
	if !strings.Contains(html, `data-invalid='true'`) {
		t.Errorf("a checked row with the danger tone must carry data-invalid\nhtml: %s", html)
	}
	if strings.Contains(html, "data-selected") {
		t.Errorf("a danger-marked row must not also carry data-selected\nhtml: %s", html)
	}

	td.SetDanger(false)
	html = row.String()
	if !strings.Contains(html, `data-selected='true'`) {
		t.Errorf("without the tone a checked row stays selected\nhtml: %s", html)
	}
	if strings.Contains(html, "data-invalid") {
		t.Errorf("without the tone no row may carry data-invalid\nhtml: %s", html)
	}
}

// THE regression, mirror of targetlist: a normal-mode highlight (Selected on
// the <li>, no selection mode) must leave the check box stateless, or the
// pencil-reveal rule fires on a plain highlight.
func TestTargetDate_NormalHighlightLeavesCheckStateless(t *testing.T) {
	td := &TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	td.Selected.Set("1")
	row := td.buildRow(Item{ID: "1", Label: "Row 1"})

	html := row.String()
	check := html[strings.Index(html, `class='targetdate__sel-check'`):]
	check = check[:strings.Index(check, "</span>")]
	if strings.Contains(check, "data-selected") || strings.Contains(check, "data-invalid") {
		t.Errorf("normal-mode highlight must leave the check box stateless\ncheck: %s", check)
	}
	if !strings.Contains(html, `data-selected='true'`) {
		t.Errorf("the row <li> must still carry the normal-mode highlight\nhtml: %s", html)
	}
}

func TestTargetDate_CheckRidesTopEndCorner(t *testing.T) {
	cssStr := (&TargetDate{Zone: time.MachineZone()}).RenderCSS().String()

	// Anchored at line start: the KeepSize primitives group also mentions
	// this class (as its last selector), so a bare substring search would
	// land on a rule that has flex-shrink/grow but no position.
	i := strings.Index(cssStr, "\n.targetdate__sel-check {")
	if i == -1 {
		t.Fatal("expected a rule for .targetdate__sel-check")
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "position: absolute") {
		t.Errorf("the check must be out of flow (position: absolute), block:\n%s", body)
	}
}

func TestTargetDate_DangerRowUsesWash(t *testing.T) {
	cssStr := (&TargetDate{Zone: time.MachineZone()}).RenderCSS().String()

	i := strings.Index(cssStr, `.targetdate__row[data-invalid="true"] {`)
	if i == -1 {
		t.Fatal("expected a rule for the invalid row state")
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "--color-danger-wash") {
		t.Errorf("a danger-marked row must use the DangerWash surface, block:\n%s", body)
	}
}

// Mirror of targetlist: glyph reveal + box colour hang off the CHECK BOX's
// own state, never the row's.
func TestTargetDate_GlyphRevealHangsOffTheCheckBox(t *testing.T) {
	cssStr := (&TargetDate{Zone: time.MachineZone()}).RenderCSS().String()

	for _, sel := range []string{
		`.targetdate__sel-check[data-invalid="true"] .targetdate__sel-check-trash`,
		`.targetdate__sel-check[data-selected="true"] .targetdate__sel-check-pencil`,
		`.targetdate__sel-check[data-selected="true"]`,
		`.targetdate__sel-check[data-invalid="true"]`,
	} {
		if !strings.Contains(cssStr, sel) {
			t.Errorf("expected a check-box-scoped rule for %s", sel)
		}
	}
	for _, sel := range []string{
		`.targetdate__row[data-selected="true"] .targetdate__sel-check-pencil`,
		`.targetdate__row[data-invalid="true"] .targetdate__sel-check-trash`,
	} {
		if strings.Contains(cssStr, sel) {
			t.Errorf("the glyph reveal must not hang off the row: %s", sel)
		}
	}
}

// The edit box must be AccentInverse (white on-primary); the delete box Danger
// (white on-danger). A near-black currentColor glyph on Accent/Inset is the
// defect this fixes.
func TestTargetDate_CheckBoxCarriesWhiteText(t *testing.T) {
	cssStr := (&TargetDate{Zone: time.MachineZone()}).RenderCSS().String()

	for _, tc := range []struct{ sel, want string }{
		{`.targetdate__sel-check[data-selected="true"]`, "--color-on-primary"},
		{`.targetdate__sel-check[data-invalid="true"]`, "--color-on-danger"},
	} {
		i := strings.Index(cssStr, tc.sel+" {")
		if i == -1 {
			t.Fatalf("expected a rule for %s", tc.sel)
		}
		body := cssStr[i:]
		if end := strings.Index(body, "}"); end != -1 {
			body = body[:end]
		}
		if !strings.Contains(body, tc.want) {
			t.Errorf("expected %s in the %s rule, block:\n%s", tc.want, tc.sel, body)
		}
		if strings.Contains(body, "--color-on-accent") {
			t.Errorf("the check must not use on-accent (near-black glyph), block:\n%s", body)
		}
	}
}

// Mirror of targetlist: the box is hidden until selection mode opens and
// revealed as a centred flex square; each glyph carries its own size.
func TestTargetDate_CheckHiddenUntilSelectionMode(t *testing.T) {
	cssStr := (&TargetDate{Zone: time.MachineZone()}).RenderCSS().String()

	// Anchored at line start: the KeepSize primitives group also mentions
	// this class (as its last selector), so a bare substring search finds no
	// display authority there.
	i := strings.Index(cssStr, "\n.targetdate__sel-check {")
	if i == -1 {
		t.Fatal("expected a rule for .targetdate__sel-check")
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "display: none") {
		t.Errorf("the check box must be hidden by default (normal mode), block:\n%s", body)
	}

	i = strings.Index(cssStr, `.targetdate[data-open="true"] .targetdate__sel-check {`)
	if i == -1 {
		t.Fatal("expected the check box to be revealed by the list's open state")
	}
	body = cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "display: flex") || !strings.Contains(body, "justify-content: center") {
		t.Errorf("the revealed box must be a centred flex box, block:\n%s", body)
	}

	for _, part := range []string{".targetdate__sel-check-trash", ".targetdate__sel-check-pencil"} {
		i := strings.Index(cssStr, part+" {")
		if i == -1 {
			t.Fatalf("expected a rule for %s", part)
		}
		body := cssStr[i:]
		if end := strings.Index(body, "}"); end != -1 {
			body = body[:end]
		}
		if !strings.Contains(body, "width:") || !strings.Contains(body, "height:") {
			t.Errorf("%s must carry its own IconBox size, block:\n%s", part, body)
		}
	}
}

// The row markup references the shared webtyp/icons glyphs, not a fixed tick.
func TestTargetDate_RenderUsesSharedIconGlyphs(t *testing.T) {
	td := &TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	row := td.buildRow(Item{ID: "1", Label: "Row 1"})

	html := row.String()
	if !strings.Contains(html, `href='#trash'`) || !strings.Contains(html, `href='#pencil'`) {
		t.Errorf("the row must reference the shared trash/pencil glyphs\nhtml: %s", html)
	}
	if strings.Contains(html, "td-check") || strings.Contains(html, "crud-delete") {
		t.Errorf("the old fixed/crudicons glyph ids must be gone\nhtml: %s", html)
	}
}
