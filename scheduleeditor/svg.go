//go:build !wasm

package scheduleeditor

import "webtyp.com/svg/sprite"

// IconSvg returns the SSR sprite for ScheduleEditor. The component itself
// declares no glyphs (its calendar lives in calendarslider, which owns its own
// IconSvg); an empty sprite satisfies the contract so ssr emits nothing.
func (e *ScheduleEditor) IconSvg() *sprite.Sprite {
	return sprite.NewSprite()
}
