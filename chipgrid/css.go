//go:build !wasm

package chipgrid

import "webtyp.com/css"

func (g *ChipGrid) RenderCSS() *css.Stylesheet { panic("not implemented") }
