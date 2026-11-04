//go:build !wasm

package targetdate_test

import (
	"strings"
	"testing"
	"webtyp.com/time"

	"webtyp.com/components/targetdate"
)

// API-level assembly tests. Row MARKUP is asserted in the internal test file,
// which can reach buildRow — Render() binds its children and SSR does not
// serialize a children binding, so an external test sees an empty <ul> and any
// assertion it makes about a row is vacuously true.
//
// The toggle BEHAVIOUR — tap order, render order, the count callback — is
// covered where it lives, in listselect/listselect_test.go: a row is marked by
// a DOM click, and there is no click to dispatch under SSR.

func TestSelectModeShowsOnTheRoot(t *testing.T) {
	// The root's Open state is the hook the stylesheet reveals the checks
	// from. Without it the mode would flip in Go and change nothing on screen.
	td := &targetdate.TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	td.SetItems([]targetdate.Item{{ID: "1", Label: "Row 1"}})

	if html := td.Render().String(); strings.Contains(html, "data-open") {
		t.Errorf("the root must not carry the open state with the mode OFF\nhtml: %s", html)
	}

	td.SetSelectMode(true)
	if html := td.Render().String(); !strings.Contains(html, "data-open='true'") {
		t.Errorf("the root must carry data-open='true' with the mode ON\nhtml: %s", html)
	}
}

func TestNothingIsCheckedUntilTheUserMarksIt(t *testing.T) {
	td := &targetdate.TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	td.SetItems([]targetdate.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})

	td.SetSelectMode(true)
	if ids := td.CheckedIDs(); len(ids) != 0 {
		t.Errorf("entering selection mode must not preselect anything, got %v", ids)
	}
}

func TestLeavingSelectModeClearsTheMarks(t *testing.T) {
	td := &targetdate.TargetDate{Zone: time.MachineZone()}
	td.Init(nil)
	td.SetItems([]targetdate.Item{{ID: "1"}, {ID: "2"}})

	td.SetSelectMode(true)
	td.SetSelectMode(false)

	if ids := td.CheckedIDs(); len(ids) != 0 {
		t.Errorf("leaving selection mode must clear the marks, got %v", ids)
	}
}
