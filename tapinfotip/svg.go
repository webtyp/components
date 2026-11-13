//go:build !wasm

package tapinfotip

import (
	"webtyp.com/icons/circleinfo"
	"webtyp.com/svg/sprite"
)

// IconSvg hands the information glyph to the sprite. The geometry stays here,
// behind !wasm; the browser only receives the id the button references.
func (t *TapInfotip) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(circleinfo.Def())
}
