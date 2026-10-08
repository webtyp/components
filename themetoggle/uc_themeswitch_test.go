//go:build wasm

package themetoggle

import (
	"testing"

	"webtyp.com/dom"
	"webtyp.com/fmt"
)

func setUp() {
	dom.LocalStorageDel(storageKey)
	dom.SetDocumentAttr("data-theme", "")
	dom.SetDocumentAttr("style", "")
}

// forceScheme stands in for the app's stylesheet: css.DefaultLight() or the
// OS preference both end up as the color-scheme <html> resolves to.
func forceScheme(scheme string) {
	dom.SetDocumentAttr("style", "color-scheme: "+scheme)
}

// With nothing saved the toggle must not impose a theme: <html> stays
// untouched and the icon shows whatever scheme is in effect.
func TestThemeToggle_Init_NoSavedValue_FollowsEffectiveScheme(t *testing.T) {
	if !dom.SupportsLightDark() {
		t.Skip("browser cannot resolve light-dark()")
	}
	for _, scheme := range []string{"light", "dark"} {
		setUp()
		forceScheme(scheme)
		ts := &ThemeToggle{}
		ts.Init(nil)

		if got := dom.GetDocumentAttr("data-theme"); got != "" {
			t.Errorf("scheme %s: Init must not write data-theme without a saved choice, got %q", scheme, got)
		}
		if got := ts.theme.Get(); got != scheme {
			t.Errorf("scheme %s: signal must follow the effective scheme, got %q", scheme, got)
		}
	}
	setUp()
}

func TestThemeToggle_Init_RestoresLight(t *testing.T) {
	setUp()
	dom.LocalStorageSet(storageKey, "light")

	ts := &ThemeToggle{}
	ts.Init(nil)

	got := dom.GetDocumentAttr("data-theme")
	if got != "light" {
		t.Errorf("expected data-theme=light, got %q", got)
	}
	if ts.theme.Get() != "light" {
		t.Errorf("expected signal=light, got %q", ts.theme.Get())
	}
}

func TestThemeToggle_Init_InvalidValue_FollowsEffectiveScheme(t *testing.T) {
	setUp()
	forceScheme("dark")
	dom.LocalStorageSet(storageKey, "xyz")

	ts := &ThemeToggle{}
	ts.Init(nil)

	if got := dom.GetDocumentAttr("data-theme"); got != "" {
		t.Errorf("an invalid saved value is no choice: data-theme must stay unset, got %q", got)
	}
	if dom.SupportsLightDark() && ts.theme.Get() != "dark" {
		t.Errorf("expected signal=dark (the effective scheme), got %q", ts.theme.Get())
	}
	setUp()
}

func TestThemeToggle_Render_Initial(t *testing.T) {
	setUp()
	forceScheme("light")
	ts := &ThemeToggle{}
	ts.Init(nil)
	el := ts.Render()
	setUp()

	// effective scheme is light → icon is ☀️
	got := el.String()
	if !fmt.Contains(got, icon(TsThemeLight)) {
		t.Errorf("expected icon %s in rendered element, got %s", icon(TsThemeLight), got)
	}
}
