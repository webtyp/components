//go:build !wasm

package targethour

import (
	"strings"
	"testing"
)

// Mirror of targetlist/targetdate's danger contract.
func TestTargetHour_DangerMarksInvalidNotSelected(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	row := th.buildRow(Item{ID: "1", Label: "Row 1", Badge: "09:00"})

	th.SetSelectMode(true)
	th.SetDanger(true)
	th.sel.Toggle("1")

	html := row.String()
	if !strings.Contains(html, `data-invalid='true'`) {
		t.Errorf("a checked row with the danger tone must carry data-invalid\nhtml: %s", html)
	}
	if strings.Contains(html, "data-selected") {
		t.Errorf("a danger-marked row must not also carry data-selected\nhtml: %s", html)
	}

	th.SetDanger(false)
	html = row.String()
	if !strings.Contains(html, `data-selected='true'`) {
		t.Errorf("without the tone a checked row stays selected\nhtml: %s", html)
	}
	if strings.Contains(html, "data-invalid") {
		t.Errorf("without the tone no row may carry data-invalid\nhtml: %s", html)
	}
}

// A normal-mode highlight must leave the check box stateless.
func TestTargetHour_NormalHighlightLeavesCheckStateless(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	th.Selected.Set("1")
	row := th.buildRow(Item{ID: "1", Label: "Row 1", Badge: "09:00"})

	html := row.String()
	check := html[strings.Index(html, `class='targethour__sel-check'`):]
	check = check[:strings.Index(check, "</span>")]
	if strings.Contains(check, "data-selected") || strings.Contains(check, "data-invalid") {
		t.Errorf("normal-mode highlight must leave the check box stateless\ncheck: %s", check)
	}
	if !strings.Contains(html, `data-selected='true'`) {
		t.Errorf("the row <li> must still carry the normal-mode highlight\nhtml: %s", html)
	}
}

func TestTargetHour_CheckRidesTopEndCorner(t *testing.T) {
	cssStr := (&TargetHour{}).RenderCSS().String()

	// Anchored at line start: the KeepSize primitives group also mentions
	// this class (as its last selector), so a bare substring search would
	// land on a rule that has flex-shrink/grow but no position.
	i := strings.Index(cssStr, "\n.targethour__sel-check {")
	if i == -1 {
		t.Fatal("expected a rule for .targethour__sel-check")
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "position: absolute") {
		t.Errorf("the check must be out of flow (position: absolute), block:\n%s", body)
	}
}

func TestTargetHour_DangerRowUsesWash(t *testing.T) {
	cssStr := (&TargetHour{}).RenderCSS().String()

	i := strings.Index(cssStr, `.targethour__row[data-invalid="true"] {`)
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

func TestTargetHour_GlyphRevealHangsOffTheCheckBox(t *testing.T) {
	cssStr := (&TargetHour{}).RenderCSS().String()

	for _, sel := range []string{
		`.targethour__sel-check[data-invalid="true"] .targethour__sel-check-trash`,
		`.targethour__sel-check[data-selected="true"] .targethour__sel-check-pencil`,
		`.targethour__sel-check[data-selected="true"]`,
		`.targethour__sel-check[data-invalid="true"]`,
	} {
		if !strings.Contains(cssStr, sel) {
			t.Errorf("expected a check-box-scoped rule for %s", sel)
		}
	}
	for _, sel := range []string{
		`.targethour__row[data-selected="true"] .targethour__sel-check-pencil`,
		`.targethour__row[data-invalid="true"] .targethour__sel-check-trash`,
	} {
		if strings.Contains(cssStr, sel) {
			t.Errorf("the glyph reveal must not hang off the row: %s", sel)
		}
	}
}

func TestTargetHour_CheckBoxCarriesWhiteText(t *testing.T) {
	cssStr := (&TargetHour{}).RenderCSS().String()

	for _, tc := range []struct{ sel, want string }{
		{`.targethour__sel-check[data-selected="true"]`, "--color-on-primary"},
		{`.targethour__sel-check[data-invalid="true"]`, "--color-on-danger"},
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

func TestTargetHour_CheckHiddenUntilSelectionMode(t *testing.T) {
	cssStr := (&TargetHour{}).RenderCSS().String()

	// Anchored at line start: the KeepSize primitives group also mentions
	// this class (as its last selector), so a bare substring search finds no
	// display authority there.
	i := strings.Index(cssStr, "\n.targethour__sel-check {")
	if i == -1 {
		t.Fatal("expected a rule for .targethour__sel-check")
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !strings.Contains(body, "display: none") {
		t.Errorf("the check box must be hidden by default (normal mode), block:\n%s", body)
	}

	i = strings.Index(cssStr, `.targethour[data-open="true"] .targethour__sel-check {`)
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

	for _, part := range []string{".targethour__sel-check-trash", ".targethour__sel-check-pencil"} {
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

func TestTargetHour_RenderUsesSharedIconGlyphs(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	row := th.buildRow(Item{ID: "1", Label: "Row 1", Badge: "09:00"})

	html := row.String()
	if !strings.Contains(html, `href='#trash'`) || !strings.Contains(html, `href='#pencil'`) {
		t.Errorf("the row must reference the shared trash/pencil glyphs\nhtml: %s", html)
	}
}
