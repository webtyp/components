//go:build !wasm

package stepaccordion

import "webtyp.com/css"

func (a *StepAccordion) RenderCSS() *css.Stylesheet { panic("not implemented") }
