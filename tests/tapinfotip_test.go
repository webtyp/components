//go:build !wasm

package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/tapinfotip"
)

// The ARIA contract: the (i) names the panel it opens, starts collapsed, and
// never submits a form it sits in.
func TestTapInfotipButtonNamesItsPanel(t *testing.T) {
	tip := &tapinfotip.TapInfotip{Help: "Document number, as written on the document."}
	button := tip.Button().String()
	panel := tip.Panel()
	if !strings.Contains(button, panel.GetID()) {
		t.Errorf("button does not reference the panel id %q:\n%s", panel.GetID(), button)
	}
}

func TestTapInfotipStartsCollapsed(t *testing.T) {
	tip := &tapinfotip.TapInfotip{Help: "x"}
	if button := tip.Button().String(); !strings.Contains(button, "aria-expanded='false'") {
		t.Errorf("a new (i) must start with aria-expanded=false:\n%s", button)
	}
}

func TestTapInfotipButtonDoesNotSubmit(t *testing.T) {
	tip := &tapinfotip.TapInfotip{Help: "x"}
	if button := tip.Button().String(); !strings.Contains(button, "type='button'") {
		t.Errorf("the (i) inside a form would submit it:\n%s", button)
	}
}

// The help text is the one the author wrote, so a form shows what the model
// declares.
func TestTapInfotipPanelCarriesHelpText(t *testing.T) {
	const help = "Uncheck to retire the patient; the record is never deleted."
	tip := &tapinfotip.TapInfotip{Help: help}
	if panel := tip.Panel().String(); !strings.Contains(panel, help) {
		t.Errorf("panel does not carry the help text:\n%s", panel)
	}
}
