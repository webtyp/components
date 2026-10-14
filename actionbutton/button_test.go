//go:build !wasm

package actionbutton

import (
	"regexp"
	"strings"
	"testing"

	. "webtyp.com/dom"
)

var (
	classRegex      = regexp.MustCompile(`\.([a-zA-Z0-9_-]+)`)
	htmlClassRegex  = regexp.MustCompile(`class='([^']*)'`)
	htmlClassRegex2 = regexp.MustCompile(`class="([^"]*)"`)
)

func TestButton_Render(t *testing.T) {
	btn := &ActionButton{
		Text:    "Click",
		Variant: Primary,
	}

	html := btn.Render().String()

	if !strings.HasPrefix(html, "<button") {
		t.Error("expected button tag, got: " + html)
	}

	// Verify classes
	if !strings.Contains(html, "actionbutton") || !strings.Contains(html, "actionbutton__primary") {
		t.Error("expected actionbutton classes, got: " + html)
	}

	// Verify text
	if !strings.Contains(html, ">Click<") {
		t.Error("expected text 'Click', got: " + html)
	}
}

// TestButton_RenderHref proves the consumer-shaped use case that motivated
// Href: an OAuth login link that must work before WASM loads and with
// JavaScript disabled, styled exactly like a primary button.
func TestButton_RenderHref(t *testing.T) {
	btn := &ActionButton{
		Text:    "Sign in with Google",
		Variant: Primary,
		Href:    "/oauth/google",
	}

	html := btn.Render().String()

	if !strings.HasPrefix(html, "<a") {
		t.Error("expected an <a> tag for a Href button, got: " + html)
	}
	if !strings.Contains(html, `href='/oauth/google'`) {
		t.Error("expected href attribute, got: " + html)
	}
	if !strings.Contains(html, "actionbutton") || !strings.Contains(html, "actionbutton__primary") {
		t.Error("expected the same actionbutton classes as the button variant, got: " + html)
	}
	if strings.Contains(html, "<button") {
		t.Error("expected no <button> tag when Href is set, got: " + html)
	}
}

// TestButton_HrefWinsOverOnClick documents the precedence: Href is for
// navigation, so it takes over even if a caller also sets OnClick.
func TestButton_HrefWinsOverOnClick(t *testing.T) {
	called := false
	btn := &ActionButton{
		Text:    "Go",
		Href:    "/somewhere",
		OnClick: func(Event) { called = true },
	}

	html := btn.Render().String()

	if !strings.HasPrefix(html, "<a") {
		t.Error("expected Href to win and render an <a> tag, got: " + html)
	}
	if called {
		t.Error("OnClick must not fire from Render when Href is set")
	}
}

func TestPairMarkupAndStylesheet(t *testing.T) {
	extractCSSClasses := func(css string) map[string]bool {
		classes := make(map[string]bool)
		matches := classRegex.FindAllStringSubmatch(css, -1)
		for _, m := range matches {
			if len(m) > 1 {
				classes[m[1]] = true
			}
		}
		return classes
	}

	extractHTMLClasses := func(html string) map[string]bool {
		classes := make(map[string]bool)
		matches1 := htmlClassRegex.FindAllStringSubmatch(html, -1)
		for _, m := range matches1 {
			if len(m) > 1 {
				for _, cls := range strings.Fields(m[1]) {
					classes[cls] = true
				}
			}
		}
		matches2 := htmlClassRegex2.FindAllStringSubmatch(html, -1)
		for _, m := range matches2 {
			if len(m) > 1 {
				for _, cls := range strings.Fields(m[1]) {
					classes[cls] = true
				}
			}
		}
		return classes
	}

	filterClasses := func(classes map[string]bool, prefix string) map[string]bool {
		filtered := make(map[string]bool)
		for cls := range classes {
			if strings.HasPrefix(cls, prefix) {
				filtered[cls] = true
			}
		}
		return filtered
	}

	cssStr := (&ActionButton{}).RenderCSS().String()
	cssClasses := filterClasses(extractCSSClasses(cssStr), "actionbutton")

	variants := []Variant{Primary, Secondary, Danger}
	htmlClasses := make(map[string]bool)
	for _, v := range variants {
		btn := &ActionButton{Text: "Button", Variant: v}
		for cls := range extractHTMLClasses(btn.Render().String()) {
			htmlClasses[cls] = true
		}
	}
	htmlClasses = filterClasses(htmlClasses, "actionbutton")

	for cls := range cssClasses {
		if !htmlClasses[cls] {
			t.Errorf("ActionButton CSS class %q does not exist in rendered HTML", cls)
		}
	}
	for cls := range htmlClasses {
		if !cssClasses[cls] {
			t.Errorf("ActionButton HTML class %q is unstyled in CSS", cls)
		}
	}
}
