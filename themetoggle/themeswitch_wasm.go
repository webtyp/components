//go:build wasm

package themetoggle

import . "webtyp.com/dom"

func (t *ThemeToggle) Init(_ Ctx) {
	t.supported = SupportsLightDark() // before the first Render — see the field's doc comment
	if LocalStorageAvailable() {
		if s, err := LocalStorageGet(storageKey); err == nil && valid(TsTheme(s)) {
			t.theme = NewString(s)           // value ready before first paint → correct icon, no flash
			SetDocumentAttr("data-theme", s) // the user's own choice, applied
			return
		}
	}
	// No choice yet: show the scheme in effect and leave <html> alone, so the
	// app's declared default (or the OS) keeps deciding.
	effective := TsThemeLight
	if DarkSchemeActive() {
		effective = TsThemeDark
	}
	t.theme = NewString(string(effective))
}

func (t *ThemeToggle) onClick() {
	next := toggle(TsTheme(t.theme.Get()))
	SetDocumentAttr("data-theme", string(next)) // applies the theme
	if LocalStorageAvailable() {
		LocalStorageSet(storageKey, string(next))
	}
	t.theme.Set(string(next)) // patches icon + labels surgically
}
