//go:build wasm

package scheduleeditor_test

import "time"

// The calendar shows months from the current one on, so a test date must move with the clock:
// fixed dates silently fall out of the rendered range (a "not selectable" assertion then passes
// without a single day to check). thu, fri, sat and sun are one Thursday-to-Sunday week in the
// first week of next month, always inside the editor's horizon.
var thu, fri, sat, sun = nextMonthWeek(time.Now())

func nextMonthWeek(now time.Time) (string, string, string, string) {
	d := time.Date(now.Year(), now.Month()+1, 1, 12, 0, 0, 0, time.Local)
	for d.Weekday() != time.Thursday {
		d = d.AddDate(0, 0, 1)
	}
	day := func(n int) string { return d.AddDate(0, 0, n).Format("2006-01-02") }
	return day(0), day(1), day(2), day(3)
}
