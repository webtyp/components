//go:build !wasm

package targethour

import (
	"strings"
	"testing"

	"webtyp.com/view"
	"webtyp.com/widget"
)

// A row with no StatusOf carries no tint state.
func TestTargetHour_NoStatusMapperNoTint(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	html := th.buildRow(Item{ID: "1", Label: "x", Badge: "09:00"}).String()
	if strings.Contains(html, string(widget.Locked.Attr().Key())+"='true'") ||
		strings.Contains(html, string(widget.Busy.Attr().Key())+"='true'") {
		t.Errorf("a row without a StatusOf mapper must carry no tint state\n%s", html)
	}
}

// StatusConfirmed -> data-locked='true'; StatusAttended -> data-busy='true';
// StatusPending -> neither.
func TestTargetHour_StatusDrivesTheTintState(t *testing.T) {
	cases := []struct {
		st        Status
		wantKey   string
		wantOther string
	}{
		{StatusConfirmed, widget.Locked.Attr().Key(), widget.Busy.Attr().Key()},
		{StatusAttended, widget.Busy.Attr().Key(), widget.Locked.Attr().Key()},
	}
	for _, c := range cases {
		th := &TargetHour{StatusOf: func(view.Item) Status { return c.st }}
		th.Init(nil)
		html := th.buildRow(Item{ID: "1", Label: "x", Badge: "09:00"}).String()
		if !strings.Contains(html, c.wantKey+"='true'") {
			t.Errorf("status %d must set %s='true'\n%s", c.st, c.wantKey, html)
		}
		if strings.Contains(html, c.wantOther+"='true'") {
			t.Errorf("status %d must NOT set %s='true'\n%s", c.st, c.wantOther, html)
		}
	}
}

// The lead renders it.LeadMain as the prominent hour.
func TestTargetHour_LeadRendersTheHour(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	html := th.buildRow(Item{ID: "1", Label: "Ana", Badge: "14:30"}).String()
	if !strings.Contains(html, "targethour__hour") || !strings.Contains(html, "14:30") {
		t.Errorf("the lead must render LeadMain as the hour\n%s", html)
	}
}
