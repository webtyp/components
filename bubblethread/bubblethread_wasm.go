//go:build wasm

package bubblethread

import . "webtyp.com/dom"

func scrollToLast(id string) {
	if el, ok := Get(id); ok && el != nil {
		el.ScrollIntoViewInstant()
	}
}
