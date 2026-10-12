//go:build !wasm

package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/stepaccordion"
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

func stateAttr(s widget.State) string { return s.Key() + "='" + s.Value() + "'" }

func threeSteps() []stepaccordion.Step {
	return []stepaccordion.Step{
		{Key: "service", Title: "Service", Body: html.P().Text("BODY-1")},
		{Key: "staff", Title: "Professional", Body: html.P().Text("BODY-2")},
		{Key: "slot", Title: "Date and time", Body: html.P().Text("BODY-3")},
	}
}

func newAccordion(steps []stepaccordion.Step) *stepaccordion.StepAccordion {
	a := &stepaccordion.StepAccordion{Steps: steps}
	a.Init(nil)
	return a
}

// sections returns the rendered markup of each step <section>, in order.
func sections(t *testing.T, a *stepaccordion.StepAccordion) []string {
	t.Helper()
	parts := strings.Split(a.Render().String(), "<section")
	return parts[1:]
}

// phase reads which of the three phases a rendered step is in.
func phase(section string) string {
	open := strings.Contains(section, stateAttr(widget.Open))
	done := strings.Contains(section, stateAttr(widget.Locked))
	locked := strings.Contains(section, stateAttr(widget.Disabled))
	switch {
	case open && !done && !locked:
		return "open"
	case done && !open && !locked:
		return "done"
	case locked && !open && !done:
		return "locked"
	}
	return "invalid"
}

func phases(t *testing.T, a *stepaccordion.StepAccordion) []string {
	t.Helper()
	var out []string
	for _, s := range sections(t, a) {
		out = append(out, phase(s))
	}
	return out
}

func wantPhases(t *testing.T, a *stepaccordion.StepAccordion, want ...string) {
	t.Helper()
	got := phases(t, a)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("phases = %v, want %v", got, want)
	}
}

func TestStepAccordion_InitFirstOpenRestLocked(t *testing.T) {
	a := newAccordion(threeSteps())
	wantPhases(t, a, "open", "locked", "locked")
	if a.Open() != "service" {
		t.Errorf("Open() = %q, want service", a.Open())
	}
}

func TestStepAccordion_ZeroStepsRendersEmptyRoot(t *testing.T) {
	a := newAccordion(nil)
	out := a.Render().String()
	if !strings.Contains(out, "class='stepaccordion'") {
		t.Errorf("expected the root, got %s", out)
	}
	if strings.Contains(out, "<section") {
		t.Errorf("expected no sections, got %s", out)
	}
	if a.Open() != "" {
		t.Errorf("Open() = %q, want empty", a.Open())
	}
}

func TestStepAccordion_RenderWithoutInit(t *testing.T) {
	a := &stepaccordion.StepAccordion{Steps: threeSteps()}
	wantPhases(t, a, "open", "locked", "locked")
}

func TestStepAccordion_AdvanceWithNothingOpenIsNoop(t *testing.T) {
	a := newAccordion(threeSteps()[:1])
	a.Advance("one")
	a.Advance("two") // nothing open any more
	wantPhases(t, a, "done")
	if !strings.Contains(sections(t, a)[0], ">one<") {
		t.Error("the second Advance must not overwrite the summary")
	}
}

func TestStepAccordion_AdvanceOpensNextAndSetsSummary(t *testing.T) {
	a := newAccordion(threeSteps())
	a.Advance("Consulta")
	wantPhases(t, a, "done", "open", "locked")
	if !strings.Contains(sections(t, a)[0], ">Consulta<") {
		t.Error("the finished step must show its summary")
	}
	if a.Open() != "staff" {
		t.Errorf("Open() = %q, want staff", a.Open())
	}
}

func TestStepAccordion_AdvanceOnLastLeavesNothingOpen(t *testing.T) {
	a := newAccordion(threeSteps())
	a.Advance("a")
	a.Advance("b")
	a.Advance("c")
	wantPhases(t, a, "done", "done", "done")
	if a.Open() != "" {
		t.Errorf("Open() = %q, want empty", a.Open())
	}
}

func TestStepAccordion_ReopenUnknownKeyIsNoop(t *testing.T) {
	a := newAccordion(threeSteps())
	a.Advance("a")
	a.Reopen("nope")
	wantPhases(t, a, "done", "open", "locked")
}

func TestStepAccordion_ReopenLockedOrOpenIsNoop(t *testing.T) {
	calls := 0
	a := newAccordion(threeSteps())
	a.OnReopen = func(string) { calls++ }
	a.Advance("a")
	a.Reopen("staff") // open
	a.Reopen("slot")  // locked
	wantPhases(t, a, "done", "open", "locked")
	if calls != 0 {
		t.Errorf("OnReopen called %d times, want 0", calls)
	}
}

func TestStepAccordion_ReopenLocksLaterStepsAndCallsOnReopen(t *testing.T) {
	var got []string
	a := newAccordion(threeSteps())
	a.OnReopen = func(key string) { got = append(got, key) }
	a.Advance("Consulta")
	a.Advance("Dra. Rojas")
	a.Reopen("service")
	wantPhases(t, a, "open", "locked", "locked")
	for i, s := range sections(t, a)[1:] {
		if strings.Contains(s, "Dra. Rojas") {
			t.Errorf("step %d kept its summary after an earlier step was reopened", i+2)
		}
	}
	if strings.Join(got, ",") != "service" {
		t.Errorf("OnReopen got %v, want [service]", got)
	}
	if a.Open() != "service" {
		t.Errorf("Open() = %q, want service", a.Open())
	}
}

func TestStepAccordion_ReopenWithoutCallback(t *testing.T) {
	a := newAccordion(threeSteps())
	a.Advance("a")
	a.Advance("b")
	a.Reopen("staff")
	wantPhases(t, a, "done", "open", "locked")
}

func TestStepAccordion_EditControlsBodyAndBodyLabelledByTitle(t *testing.T) {
	a := newAccordion(threeSteps()[:1])
	s := sections(t, a)[0]
	if !strings.Contains(s, "aria-expanded='false'") {
		t.Error("the Edit button must declare aria-expanded='false'")
	}
	controls := attrValue(s, "aria-controls")
	labelled := attrValue(s, "aria-labelledby")
	if controls == "" || labelled == "" {
		t.Fatalf("missing aria-controls (%q) or aria-labelledby (%q)", controls, labelled)
	}
	if !strings.Contains(s, "id='"+controls+"'") {
		t.Errorf("aria-controls %q names no element of the step", controls)
	}
	if !strings.Contains(s, "id='"+labelled+"'") {
		t.Errorf("aria-labelledby %q names no element of the step", labelled)
	}
}

func TestStepAccordion_TitleAndEditTranslated(t *testing.T) {
	if err := lang.Load([]byte(`{"languages":["es"],"keys":{"Edit":["Editar"],"Service":["Servicio"]}}`)); err != nil {
		t.Fatal(err)
	}
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)
	s := sections(t, newAccordion(threeSteps()[:1]))[0]
	if !strings.Contains(s, ">Editar<") || !strings.Contains(s, ">Servicio<") {
		t.Errorf("expected translated title and Edit, got %s", s)
	}
}

func TestStepAccordion_StepNumbersAndBodies(t *testing.T) {
	secs := sections(t, newAccordion(threeSteps()))
	for i, want := range []string{">1<", ">2<", ">3<"} {
		if !strings.Contains(secs[i], want) {
			t.Errorf("step %d: missing badge %s", i+1, want)
		}
	}
	if !strings.Contains(secs[0], "BODY-1") || !strings.Contains(secs[2], "BODY-3") {
		t.Error("every step renders its body (visibility is the stylesheet's job)")
	}
}

// The body shows only on the open step; summary and Edit only on a finished
// one. Each reveal hangs off the step's state.
func TestStepAccordion_StylesheetRevealsByPhase(t *testing.T) {
	out := (&stepaccordion.StepAccordion{}).RenderCSS().String()
	for _, want := range []string{
		".stepaccordion__step[" + widget.Open.Key() + "=\"" + widget.Open.Value() + "\"] .stepaccordion__body",
		".stepaccordion__step[" + widget.Locked.Key() + "=\"" + widget.Locked.Value() + "\"] .stepaccordion__summary",
		".stepaccordion__step[" + widget.Locked.Key() + "=\"" + widget.Locked.Value() + "\"] .stepaccordion__edit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stylesheet lacks %s", want)
		}
	}
}

func TestStepAccordion_RenderCSS(t *testing.T) {
	sheet := (&stepaccordion.StepAccordion{}).RenderCSS()
	if sheet == nil || sheet.String() == "" {
		t.Fatal("empty stylesheet")
	}
}

// attrValue returns the first value of name='…' in s.
func attrValue(s, name string) string {
	i := strings.Index(s, name+"='")
	if i < 0 {
		return ""
	}
	rest := s[i+len(name)+2:]
	return rest[:strings.Index(rest, "'")]
}

var _ dom.Component = (*stepaccordion.StepAccordion)(nil)
