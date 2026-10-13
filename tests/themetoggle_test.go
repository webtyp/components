//go:build !wasm

package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/themetoggle"
	"webtyp.com/components/usermenu"
	"webtyp.com/html"
)

// ruleBlock returns the declarations of the first standalone `selector {` rule.
func ruleBlock(cssStr, selector string) string {
	i := strings.Index(cssStr, selector+" {")
	if i == -1 {
		return ""
	}
	body := cssStr[i:]
	return body[:strings.Index(body, "}")]
}

// The three options are always visible: a segmented control that hides the
// unselected ones is a single inert icon.
func TestThemeToggleOptionsAllVisible(t *testing.T) {
	ts := &themetoggle.ThemeToggle{}
	block := ruleBlock(ts.RenderCSS().String(), ".themetoggle__option")
	if block == "" {
		t.Fatal("no base rule for .themetoggle__option")
	}
	if strings.Contains(block, "display: none") {
		t.Errorf("unselected options are hidden:\n%s", block)
	}
}

// The glyph wears its own part, not the option's: sharing the class gave the
// <svg> the button's box and state rules.
func TestThemeToggleIconHasOwnPart(t *testing.T) {
	ts := &themetoggle.ThemeToggle{}
	ts.Init(nil)
	markup := ts.Render().String()
	if strings.Count(markup, "themetoggle__option") != 3 {
		t.Errorf("want exactly 3 elements with the option class, markup:\n%s", markup)
	}
	if strings.Count(markup, "themetoggle__icon") != 3 {
		t.Errorf("want 3 glyphs with the icon class, markup:\n%s", markup)
	}
	if strings.Count(markup, "role='radio'") != 3 {
		t.Errorf("want 3 radio options, markup:\n%s", markup)
	}
}

// Preferences sit between the roles and the app's actions.
func TestUserMenuPreferencesBeforeActions(t *testing.T) {
	m := &usermenu.UserMenu{
		Name:        "Ada",
		Roles:       []string{"Admin"},
		Preferences: html.Span().Text("PREFS"),
		Actions:     html.Span().Text("ACTIONS"),
	}
	m.Init(nil)
	markup := m.Render().String()
	p, a := strings.Index(markup, "usermenu__prefs"), strings.Index(markup, "usermenu__actions")
	if p == -1 || a == -1 || p > a {
		t.Errorf("prefs must render before actions (prefs=%d actions=%d)", p, a)
	}

	m2 := &usermenu.UserMenu{Name: "Ada"}
	m2.Init(nil)
	if strings.Contains(m2.Render().String(), "usermenu__prefs") {
		t.Error("prefs block rendered without Preferences")
	}
}
