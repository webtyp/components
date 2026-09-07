//go:build !wasm

package components

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"testing"
	"webtyp.com/components/actionbutton"
	"webtyp.com/components/calendarslider"
	"webtyp.com/components/contentcard"
	"webtyp.com/components/datatable"
	"webtyp.com/components/fieldset"
	"webtyp.com/components/herobanner"
	"webtyp.com/components/infobar"
	"webtyp.com/components/modaldialog"
	"webtyp.com/components/scheduleeditor"
	"webtyp.com/components/searchbar"
	"webtyp.com/components/selectsearch"
	"webtyp.com/components/sitenav"
	"webtyp.com/components/statgrid"
	"webtyp.com/components/targethour"
	"webtyp.com/components/targetlist"
	"webtyp.com/components/themetoggle"
	"webtyp.com/css"
	"webtyp.com/widget"
)

// Allowed CSS variables in webtyp.com/css v0.2.0 (and widget/style)
var allowedVars = map[string]bool{
	"--color-primary":      true,
	"--color-on-primary":   true,
	"--color-secondary":    true,
	"--color-on-secondary": true,
	"--color-success":      true,
	"--color-on-success":   true,
	"--color-error":        true,
	"--color-on-error":     true,

	"--color-background":     true,
	"--color-surface":        true,
	"--color-surface-sunken": true,
	"--color-on-surface":     true,
	"--color-outline":        true,
	"--color-muted":          true,
	"--color-hover":          true,
	"--color-selection":      true,
	"--color-on-selection":   true,
	"--color-disabled":       true,
	"--color-on-disabled":    true,

	"--color-background-light":     true,
	"--color-background-dark":      true,
	"--color-surface-light":        true,
	"--color-surface-dark":         true,
	"--color-surface-sunken-light": true,
	"--color-surface-sunken-dark":  true,
	"--color-on-surface-light":     true,
	"--color-on-surface-dark":      true,
	"--color-outline-light":        true,
	"--color-outline-dark":         true,
	"--color-muted-light":          true,
	"--color-muted-dark":           true,
	"--color-hover-light":          true,
	"--color-hover-dark":           true,
	"--color-selection-light":      true,
	"--color-selection-dark":       true,
	"--color-on-selection-light":   true,
	"--color-on-selection-dark":    true,
	"--color-disabled-light":       true,
	"--color-disabled-dark":        true,
	"--color-on-disabled-light":    true,
	"--color-on-disabled-dark":     true,

	"--color-background-hover": true,
	"--color-surface-hover":    true,
	"--color-primary-hover":    true,
	"--color-secondary-hover":  true,
	"--color-selection-hover":  true,
	"--color-success-hover":    true,
	"--color-error-hover":      true,
	"--color-muted-hover":      true,

	"--color-background-focus": true,
	"--color-surface-focus":    true,
	"--color-primary-focus":    true,
	"--color-secondary-focus":  true,
	"--color-selection-focus":  true,

	"--color-background-press": true,
	"--color-surface-press":    true,
	"--color-primary-press":    true,
	"--color-secondary-press":  true,
	"--color-selection-press":  true,

	"--text-xs":   true,
	"--text-sm":   true,
	"--text-base": true,
	"--text-lg":   true,
	"--text-xl":   true,
	"--text-2xl":  true,

	"--leading-tight":       true,
	"--leading-normal":      true,
	"--leading-relaxed":     true,
	"--font-weight-regular": true,
	"--font-weight-medium":  true,
	"--font-weight-bold":    true,
	"--tracking-tight":      true,
	"--tracking-normal":     true,
	"--tracking-wide":       true,
	"--space-0":             true,
	"--space-1":             true,
	"--space-2":             true,
	"--space-3":             true,
	"--space-4":             true,
	"--space-6":             true,
	"--space-8":             true,
	"--space-12":            true,
	"--radius-sm":           true,
	"--radius-md":           true,
	"--radius-lg":           true,
	"--radius-full":         true,
	"--shadow-sm":           true,
	"--shadow-md":           true,
	"--shadow-lg":           true,
	"--shadow-xl":           true,
	"--duration-fast":       true,
	"--duration-base":       true,
	"--duration-slow":       true,
	"--ease-in":             true,
	"--ease-out":            true,
	"--ease-in-out":         true,
	"--z-base":              true,
	"--z-dropdown":          true,
	"--z-sticky":            true,
	"--z-modal":             true,
	"--z-toast":             true,
	"--z-tooltip":           true,
	"--bp-sm":               true,
	"--bp-md":               true,
	"--bp-lg":               true,
	"--bp-xl":               true,
	"--max-w-prose":         true,
	"--max-w-content":       true,
	"--max-w-screen":        true,
}

var (
	hexColorRegex = regexp.MustCompile(`#[0-9a-fA-F]{3,6}\b`)
	rgbColorRegex = regexp.MustCompile(`(rgb|rgba|hsl|hsla)\(`)
	viewportRegex = regexp.MustCompile(`\d+(vw|vh|vmin|vmax)`)
	varRegex      = regexp.MustCompile(`var\((--[a-zA-Z0-9-]+)`)
)

func TestConformance(t *testing.T) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "docs" || info.Name() == ".git" || info.Name() == "web" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Check for forbidden imports: webtyp.com/css is only allowed in css.go files
		for _, imp := range node.Imports {
			if imp.Path != nil && imp.Path.Value == `"webtyp.com/css"` && !strings.HasSuffix(path, "css.go") {
				t.Errorf("%s: imports forbidden package \"webtyp.com/css\"", path)
			}
		}

		// Check css.go structure for full components (excluding helper packages like listgap or listselect)
		if strings.HasSuffix(path, "css.go") && !strings.Contains(path, "listgap") && !strings.Contains(path, "listselect") {
			var hasStyleMethod bool
			var renderCSSCount int
			for _, decl := range node.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Recv != nil && len(fn.Recv.List) > 0 {
						if fn.Name.Name == "Style" {
							hasStyleMethod = true
						}
						if fn.Name.Name == "RenderCSS" {
							renderCSSCount++
						}
					}
				}
			}
			if hasStyleMethod {
				t.Errorf("%s: declares Style(); the SSR CSS entry point is RenderCSS() *css.Stylesheet", path)
			}
			if renderCSSCount != 1 {
				t.Errorf("%s: expected exactly one RenderCSS() method, got %d", path, renderCSSCount)
			}
		}

		ast.Inspect(node, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			// 1. Cero literales de color / Viewport / Vars / RawRule / Media en BasicLit (strings)
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				val := lit.Value
				// Strip quotes
				if len(val) >= 2 {
					val = val[1 : len(val)-1]
				}

				// Skip inline SVGs or Base64 strings (these contain '#' and are not color literals)
				if !strings.Contains(val, "data:image/") && !strings.Contains(val, "PHN2Z") {
					// Check color literals
					if hexColorRegex.MatchString(val) {
						t.Errorf("%s: contains hex color literal %q", path, val)
					}
					if rgbColorRegex.MatchString(val) {
						t.Errorf("%s: contains rgb/hsl color literal %q", path, val)
					}
					// Check viewport units
					if viewportRegex.MatchString(val) {
						t.Errorf("%s: contains viewport unit %q", path, val)
					}
					// Check literal color names (e.g., exact match for simple color names)
					for _, colorName := range []string{"white", "black", "red", "blue", "green", "yellow"} {
						if val == colorName {
							t.Errorf("%s: contains forbidden color name %q", path, val)
						}
					}
				}

				// Check var(--...) exists in allowed vars
				matches := varRegex.FindAllStringSubmatch(val, -1)
				for _, match := range matches {
					if len(match) > 1 {
						varName := match[1]
						if !allowedVars[varName] {
							t.Errorf("%s: references non-catalog variable %q", path, varName)
						}
					}
				}
			}

			// 2. Cero RawRule / Media / RootCSS / Theme / Declare identifiers
			if ident, ok := n.(*ast.Ident); ok {
				if ident.Name == "Raw" {
					t.Errorf("%s: uses forbidden css.Raw (escape hatch — report the gap upstream in widget/style)", path)
				}
				if ident.Name == "RawItem" {
					t.Errorf("%s: uses forbidden css.RawItem (escape hatch — report the gap upstream in widget/style)", path)
				}
				if ident.Name == "RawRule" {
					t.Errorf("%s: uses forbidden RawRule", path)
				}
				if ident.Name == "Media" {
					t.Errorf("%s: uses forbidden Media", path)
				}
				if ident.Name == "RootCSS" {
					t.Errorf("%s: uses forbidden RootCSS (only reserved for app / webtyp/dom)", path)
				}
				if ident.Name == "Theme" {
					t.Errorf("%s: uses forbidden Theme (only reserved for app / webtyp/dom)", path)
				}
				if ident.Name == "Declare" {
					t.Errorf("%s: uses forbidden Declare (only reserved for app / webtyp/dom)", path)
				}
			}

			// 3. Cero constantes de clase escritas a mano (Class = "...")
			if valueSpec, ok := n.(*ast.ValueSpec); ok {
				isClassType := false
				if valueSpec.Type != nil {
					if ident, ok := valueSpec.Type.(*ast.Ident); ok && ident.Name == "Class" {
						isClassType = true
					} else if sel, ok := valueSpec.Type.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "css" && sel.Sel.Name == "Class" {
							isClassType = true
						}
					}
				}
				if isClassType {
					for _, valExpr := range valueSpec.Values {
						if _, ok := valExpr.(*ast.BasicLit); ok {
							// Forbid hand-written Class assignment
							t.Errorf("%s: hand-written Class constant %q detected", path, valueSpec.Names[0].Name)
						}
					}
				}
			}

			return true
		})

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestEveryPackageEmits(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderCSS panicked: %v", r)
		}
	}()

	components := []interface {
		RenderCSS() *css.Stylesheet
	}{
		&actionbutton.ActionButton{},
		&calendarslider.CalendarSlider{},
		&contentcard.ContentCard{},
		&datatable.DataTable{},
		&fieldset.Fieldset{},
		&herobanner.HeroBanner{},
		&infobar.InfoBar{},
		&modaldialog.ModalDialog{},
		&scheduleeditor.ScheduleEditor{},
		&searchbar.SearchBar{},
		&selectsearch.SelectSearch{},
		&sitenav.SiteNav{},
		&statgrid.StatGrid{},
		&targethour.TargetHour{},
		&targetlist.TargetList{},
		&themetoggle.ThemeToggle{},
	}

	for _, c := range components {
		name := fmt.Sprintf("%T", c)
		t.Run(name, func(t *testing.T) {
			sheet := c.RenderCSS()
			if sheet == nil {
				t.Error("RenderCSS returned nil")
			}
		})
	}
}

func TestKindAllowsEveryState(t *testing.T) {
	packageComponents := map[string]interface {
		WidgetKind() widget.Kind
	}{
		"actionbutton":   &actionbutton.ActionButton{},
		"calendarslider": &calendarslider.CalendarSlider{},
		"contentcard":    &contentcard.ContentCard{},
		"datatable":      &datatable.DataTable{},
		"fieldset":       &fieldset.Fieldset{},
		"herobanner":     &herobanner.HeroBanner{},
		"infobar":        &infobar.InfoBar{},
		"modaldialog":    &modaldialog.ModalDialog{},
		"scheduleeditor": &scheduleeditor.ScheduleEditor{},
		"searchbar":      &searchbar.SearchBar{},
		"selectsearch":   &selectsearch.SelectSearch{},
		"sitenav":        &sitenav.SiteNav{},
		"statgrid":       &statgrid.StatGrid{},
		"targethour":     &targethour.TargetHour{},
		"targetlist":     &targetlist.TargetList{},
		"themetoggle":    &themetoggle.ThemeToggle{},
	}

	stateMap := map[string]widget.State{
		"Selected": widget.Selected,
		"Disabled": widget.Disabled,
		"Locked":   widget.Locked,
		"Invalid":  widget.Invalid,
		"Busy":     widget.Busy,
		"Open":     widget.Open,
		"Current":  widget.Current,
	}

	for pkg, comp := range packageComponents {
		t.Run(pkg, func(t *testing.T) {
			path := filepath.Join(pkg, "css.go")
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}

			kind := comp.WidgetKind()

			ast.Inspect(node, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if this is a call to a function/method
				var funcName string
				if ident, ok := call.Fun.(*ast.Ident); ok {
					funcName = ident.Name
				} else if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					funcName = sel.Sel.Name
				}

				if funcName != "When" && funcName != "RevealedBy" {
					return true
				}

				// The state argument is the first argument for When, or the only argument for RevealedBy
				if len(call.Args) == 0 {
					return true
				}
				arg := call.Args[0]

				// Look for widget.State
				if sel, ok := arg.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "widget" {
						stateName := sel.Sel.Name
						if stateVal, ok := stateMap[stateName]; ok {
							if !kind.Allows(stateVal) {
								t.Errorf("%s: part/rule uses state %q which is not allowed by its kind %q", path, stateName, kind.String())
							}
						}
					}
				}
				return true
			})
		})
	}
}

func TestNoRemovedSymbols(t *testing.T) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "docs" || info.Name() == ".git" || info.Name() == "web" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "css.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		str := string(content)

		forbidden := []string{
			"style.On(",
			"style.Of(",
			"Hidden()",
			"Shown()",
			"Above()",
			"Scrim()",
			"Cover()",
			"Fixed()",
			"Scrolls()",
			"style.Sunken",
		}

		for _, term := range forbidden {
			if strings.Contains(str, term) {
				t.Errorf("%s: contains forbidden term %q", path, term)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoHoverCuePairs(t *testing.T) {
	cueRe := regexp.MustCompile(`Cue\(widget\.(Hover|Focus),\s*([^,\n]+),`)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "docs" || info.Name() == ".git" || info.Name() == "web" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "css.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hovered := map[string]bool{}
		focused := map[string]bool{}
		for _, m := range cueRe.FindAllStringSubmatch(string(content), -1) {
			part := strings.TrimSpace(m[2])
			if m[1] == "Hover" {
				hovered[part] = true
			} else {
				focused[part] = true
			}
		}
		for part := range hovered {
			if !focused[part] {
				t.Errorf("%s: Cue(widget.Hover, %s, ...) has no matching Cue(widget.Focus, %s, ...) -- a keyboard user tabbing to it gets none of the preview a mouse hover gives", path, part, part)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
