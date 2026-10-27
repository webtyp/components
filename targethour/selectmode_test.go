//go:build !wasm

package targethour_test

import (
	"strings"
	"testing"

	"webtyp.com/components/targethour"
	"webtyp.com/view"
)

type listViewContract interface {
	SetItems(items []view.Item)
	Items() []view.Item
	Count() int
	SetSelectMode(on bool)
	SetDanger(on bool)
	OnCheckedChange(fn func(int))
	CheckedIDs() []string
}

var _ listViewContract = (*targethour.TargetHour)(nil)

func TestSelectModeShowsOnTheRoot(t *testing.T) {
	th := &targethour.TargetHour{}
	th.Init(nil)
	th.SetItems([]targethour.Item{{ID: "1", Label: "Row 1", Badge: "09:00"}})

	if html := th.Render().String(); strings.Contains(html, "data-open") {
		t.Errorf("the root must not carry the open state with the mode OFF\nhtml: %s", html)
	}

	th.SetSelectMode(true)
	if html := th.Render().String(); !strings.Contains(html, "data-open='true'") {
		t.Errorf("the root must carry data-open='true' with the mode ON\nhtml: %s", html)
	}
}

func TestNothingIsCheckedUntilTheUserMarksIt(t *testing.T) {
	th := &targethour.TargetHour{}
	th.Init(nil)
	th.SetItems([]targethour.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})

	th.SetSelectMode(true)
	if ids := th.CheckedIDs(); len(ids) != 0 {
		t.Errorf("entering selection mode must not preselect anything, got %v", ids)
	}
}

func TestLeavingSelectModeClearsTheMarks(t *testing.T) {
	th := &targethour.TargetHour{}
	th.Init(nil)
	th.SetItems([]targethour.Item{{ID: "1"}, {ID: "2"}})

	th.SetSelectMode(true)
	th.SetSelectMode(false)

	if ids := th.CheckedIDs(); len(ids) != 0 {
		t.Errorf("leaving selection mode must clear the marks, got %v", ids)
	}
}
