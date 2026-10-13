//go:build wasm

package themetoggle

import (
	"testing"
	"strings"

	"webtyp.com/dom"
)

func setUp() {
	dom.LocalStorageDel(storageKey)
	dom.SetDocumentAttr("data-theme", "")
	dom.SetDocumentAttr("style", "")
}

func forceScheme(scheme string) {
	dom.SetDocumentAttr("style", "color-scheme: "+scheme)
}

func TestThemeToggle_Init_FallbackSystem(t *testing.T) {
	setUp()
	ts := &ThemeToggle{}
	ts.Init(nil)

	if got := dom.GetDocumentAttr("data-theme"); got != "" {
		t.Errorf("Init must not write data-theme without a saved choice, got %q", got)
	}
	if got := ts.theme.Get(); got != string(TsThemeSystem) {
		t.Errorf("expected signal to default to system, got %q", got)
	}
	setUp()
}

func TestThemeToggle_Init_RestoresSaved(t *testing.T) {
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

func TestThemeToggle_Render_Initial(t *testing.T) {
	setUp()
	ts := &ThemeToggle{}
	ts.Init(nil)
	el := ts.Render()
	setUp()

	got := el.String()
	if !strings.Contains(got, "role='radiogroup'") {
		t.Errorf("expected radiogroup in rendered element, got %s", got)
	}
}
