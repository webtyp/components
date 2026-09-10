package themetoggle

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NameThemeToggle is the widget name.
const NameThemeToggle = widget.Name("themetoggle")

const (
	PartButton = widget.Part("button")
)

var (
	clsTsBtn = NameThemeToggle.Root()
)

// storageKey identifica la entrada de localStorage del componente.
const storageKey = "webtyp-themeswitch"

// TsTheme representa el estado de tema del componente.
type TsTheme string

const (
	TsThemeDark  TsTheme = "dark"
	TsThemeLight TsTheme = "light"
)

// defaultTheme es el tema por defecto cuando no hay nada guardado en
// localStorage (o en SSR, donde no hay localStorage).
const defaultTheme = TsThemeLight

// ThemeToggle es un botón flotante que alterna entre dark y light.
// Restaura automáticamente el tema guardado en localStorage al montarse.
type ThemeToggle struct {
	Element
	theme *SignalString // "dark" | "light"

	// supported es falso solo cuando el navegador no puede resolver
	// light-dark() (ver dom.SupportsLightDark): en ese caso el toggle no
	// tiene nada que hacer — el color ya está fijo en modo claro por el
	// fallback de webtyp/css — y un botón que no responde al pulsarlo se
	// lee como un bug, no como una limitación del dispositivo. Calculado una
	// sola vez en Init, antes del primer Render (ver themeswitch_wasm.go).
	// SSR (themeswitch_backend.go) no puede saberlo — no hay navegador — así
	// que asume true; es el WASM real quien corrige si hace falta.
	supported bool
}

func (t *ThemeToggle) WidgetName() widget.Name { return NameThemeToggle }
func (t *ThemeToggle) WidgetKind() widget.Kind { return widget.Region }

func (t *ThemeToggle) Render() *Element {
	if !t.supported {
		return Span()
	}

	// labelSig is used twice (title + aria-label) → a named shared computed. Auto-tracked: no deps list.
	labelSig := DeriveString(func() string { return label(TsTheme(t.theme.Get())) })

	return Button().
		BindTextFunc(func() string { return icon(TsTheme(t.theme.Get())) }). // computed icon, auto-tracked
		BindAttr("title", labelSig).
		BindAttr("aria-label", labelSig).
		Set(clsTsBtn.AsAttr()).
		OnClick(func(Event) {
			t.onClick()
		})
}

// toggle alterna entre los 2 estados. Switch (no map) — TinyGo.
func toggle(current TsTheme) TsTheme {
	switch current {
	case TsThemeDark:
		return TsThemeLight
	default:
		return TsThemeDark
	}
}

// icon retorna el símbolo visible del botón. Switch (no map) — TinyGo.
func icon(theme TsTheme) string {
	switch theme {
	case TsThemeDark:
		return "🌙"
	default:
		return "☀️" // U+2600 + U+FE0F (variation selector) forces emoji presentation, same size as 🌙
	}
}

// label retorna el nombre del modo actual (usado como tooltip y aria-label).
func label(theme TsTheme) string {
	switch theme {
	case TsThemeDark:
		return "dark"
	default:
		return "light"
	}
}

func valid(t TsTheme) bool {
	return t == TsThemeDark || t == TsThemeLight
}
