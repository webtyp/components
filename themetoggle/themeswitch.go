package themetoggle

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt/lang"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameThemeToggle is the widget name.
const NameThemeToggle = widget.Name("themetoggle")

const (
	PartOption = widget.Part("option")
)

var (
	clsTsBtn = NameThemeToggle.Root()
)

// storageKey identifica la entrada de localStorage del componente.
const storageKey = "webtyp-themeswitch"

// TsTheme representa el estado de tema del componente.
type TsTheme string

const (
	TsThemeSystem TsTheme = "system"
	TsThemeDark   TsTheme = "dark"
	TsThemeLight  TsTheme = "light"
)

const (
	IconSystem = svg.Icon("themetoggle-system")
	IconLight  = svg.Icon("themetoggle-light")
	IconDark   = svg.Icon("themetoggle-dark")
)

// ThemeToggle es un control segmentado que alterna entre system, light y dark.
//
// Al montarse restaura la elección guardada en localStorage. Sin elección no
// impone ningún tema: deja <html> como está y muestra el esquema en efecto —
// el que declara la app (css.DefaultLight) o, si no declara nada, el del SO.
// Solo un clic escribe data-theme.
type ThemeToggle struct {
	Element
	theme *SignalString // "system" | "dark" | "light"

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
func (t *ThemeToggle) WidgetKind() widget.Kind { return widget.Tabs }

func (t *ThemeToggle) Render() *Element {
	if !t.supported {
		return Span()
	}

	optClass := string(NameThemeToggle.Class(PartOption))

	return Div().
		Set(clsTsBtn.AsAttr()).
		Attr("role", "radiogroup").
		Attr("aria-label", "Theme").
		Child(
			t.renderOption(TsThemeSystem, "System", IconSystem, optClass),
		).
		Child(
			t.renderOption(TsThemeLight, "Light", IconLight, optClass),
		).
		Child(
			t.renderOption(TsThemeDark, "Dark", IconDark, optClass),
		)
}

func (t *ThemeToggle) renderOption(theme TsTheme, labelKey string, icon svg.Icon, optClass string) *Element {
	labelSig := DeriveString(func() string { return lang.Translate(labelKey).String() })
	isActiveSig := DeriveString(func() string {
		if t.theme.Get() == string(theme) {
			return "true"
		}
		return "false"
	})

	return Button().
		Set(widget.Class(optClass).AsAttr()).
		Attr("role", "radio").
		BindAttr("aria-checked", isActiveSig).
		BindAttr("title", labelSig).
		BindAttr("aria-label", labelSig).
		BindStateFunc(widget.Current, func() bool { return t.theme.Get() == string(theme) }).
		Child(icon.Render(optClass)).
		OnClick(func(Event) {
			t.selectTheme(theme)
		})
}

func valid(t TsTheme) bool {
	return t == TsThemeSystem || t == TsThemeDark || t == TsThemeLight
}
