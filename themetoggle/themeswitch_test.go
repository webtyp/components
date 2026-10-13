//go:build !wasm

package themetoggle

import (
	"strings"
	"testing"
)

func TestValid(t *testing.T) {
	if !valid(TsThemeDark) {
		t.Error("valid(TsThemeDark) should be true")
	}
	if !valid(TsThemeLight) {
		t.Error("valid(TsThemeLight) should be true")
	}
	if !valid(TsThemeSystem) {
		t.Error("valid(TsThemeSystem) should be true")
	}
	if valid("invalid") {
		t.Error("valid(\"invalid\") should be false")
	}
}

func TestThemeToggle_Render(t *testing.T) {
	tt := &ThemeToggle{}
	tt.Init(nil)

	html := tt.Render().String()

	if !strings.Contains(html, "role='radiogroup'") {
		t.Error("expected role='radiogroup' on root")
	}

	countRadio := strings.Count(html, "role='radio'")
	if countRadio != 3 {
		t.Errorf("expected 3 options with role='radio', got %d", countRadio)
	}

	countChecked := strings.Count(html, "aria-checked='true'")
	if countChecked != 1 {
		t.Errorf("expected exactly 1 aria-checked='true', got %d", countChecked)
	}

	if strings.Contains(html, "☀️") || strings.Contains(html, "🌙") {
		t.Error("markup contains emoji characters, expected SVG")
	}
}

func TestRenderCSS_NotEmpty(t *testing.T) {
	ts := &ThemeToggle{}
	css := ts.RenderCSS()
	if css.String() == "" {
		t.Error("Stylesheet() returned empty string")
	}
	if strings.Contains(css.String(), "position: fixed") || strings.Contains(css.String(), "position:fixed") {
		t.Error("stylesheet contains position: fixed")
	}
}
