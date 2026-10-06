//go:build !wasm

// Root-level test (justified): Public API integration for multiple component packages.
package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/actionbutton"
	"webtyp.com/components/bubblethread"
	"webtyp.com/components/composebar"
	"webtyp.com/components/decktabs"
	"webtyp.com/components/inboxlist"
	"webtyp.com/components/modaldialog"
	"webtyp.com/components/presencelist"
	"webtyp.com/components/searchbar"
	"webtyp.com/components/segmentedcontrol"
	"webtyp.com/components/selectsearch"
	"webtyp.com/components/stepindicator"
)

func TestFixedText_BackendPassthrough(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string
	}{
		{
			name: "SearchBar",
			html: (&searchbar.SearchBar{Placeholder: "Search users"}).Render().String(),
			want: "Search users",
		},
		{
			name: "ComposeBar.Placeholder",
			html: (&composebar.ComposeBar{Placeholder: "Write message", OnSend: func(string) {}}).Render().String(),
			want: "Write message",
		},
		{
			name: "ComposeBar.SendLabel",
			html: (&composebar.ComposeBar{SendLabel: "Send IT", OnSend: func(string) {}}).Render().String(),
			want: "Send IT",
		},
		{
			name: "SelectSearch.Placeholder",
			html: (&selectsearch.SelectSearch{Placeholder: "Select a user"}).Render().String(),
			want: "Select a user",
		},
		{
			name: "BubbleThread.Empty",
			html: (&bubblethread.BubbleThread{Empty: "No messages yet"}).Render().String(),
			want: "No messages yet",
		},
		{
			name: "InboxList.Empty",
			html: (&inboxlist.InboxList{Empty: "Your inbox is empty"}).Render().String(),
			want: "Your inbox is empty",
		},
		{
			name: "PresenceList.Empty",
			html: (&presencelist.PresenceList{Empty: "Nobody is online"}).Render().String(),
			want: "Nobody is online",
		},
		{
			name: "ModalDialog.Title",
			html: func() string {
				m := &modaldialog.ModalDialog{Title: "Confirm Action"}
				m.Init(nil)
				m.Open()
				return m.Render().String()
			}(),
			want: "Confirm Action",
		},
		{
			name: "DeckTabs.Label",
			html: (&decktabs.DeckTabs{Label: "Main Navigation", Items: []decktabs.Item{{ID: "1", Label: "Tab 1"}}}).Render().String(),
			want: "Main Navigation",
		},
		{
			name: "DeckTabs.Item.Label",
			html: (&decktabs.DeckTabs{Items: []decktabs.Item{{ID: "1", Label: "First Tab"}}}).Render().String(),
			want: "First Tab",
		},
		{
			name: "StepIndicator.Step.Label",
			html: (&stepindicator.StepIndicator{Steps: []stepindicator.Step{{Key: "1", Label: "Step One"}}}).Render().String(),
			want: "Step One",
		},
		{
			name: "SegmentedControl.Option.Label",
			html: (&segmentedcontrol.SegmentedControl{Options: []segmentedcontrol.Option{{Value: "1", Label: "Option One"}}}).Render().String(),
			want: "Option One",
		},
		{
			name: "ActionButton.Text",
			html: (&actionbutton.ActionButton{Text: "Click Me"}).Render().String(),
			want: "Click Me",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.html, tc.want) {
				t.Errorf("expected HTML to contain %q, got: %s", tc.want, tc.html)
			}
		})
	}
}
