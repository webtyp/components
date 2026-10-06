//go:build !wasm
package components_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type LangFile struct {
	Default   string              `json:"default,omitempty"`
	Languages []string            `json:"languages"`
	Keys      map[string][]string `json:"keys"`
}

func TestLangJsonIsValid(t *testing.T) {
	// Look for lang.json in the current working dir if tests run from root
	path := "lang.json"
	data, err := os.ReadFile(path)
	if err != nil {
		path = filepath.Join("..", "lang.json")
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read lang.json: %v", err)
		}
	}

	var lf LangFile
	if err := json.Unmarshal(data, &lf); err != nil {
		t.Fatalf("lang.json is not valid JSON: %v", err)
	}

	if len(lf.Languages) == 0 {
		t.Errorf("expected languages to have at least one entry")
	}

	if len(lf.Keys) == 0 {
		t.Errorf("expected at least one key in keys")
	}

	for k, list := range lf.Keys {
		if len(list) != len(lf.Languages) {
			t.Errorf("key %q has %d translations, expected %d to match languages array", k, len(list), len(lf.Languages))
		}
	}
}
