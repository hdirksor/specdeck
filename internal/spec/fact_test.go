package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestParseFact_Boolean(t *testing.T) {
	f := writeTempFile(t, `
name: is-logged-in
type: boolean
`)
	fact, err := spec.ParseFact(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.Name != "is-logged-in" {
		t.Errorf("expected name 'is-logged-in', got %q", fact.Name)
	}
	if fact.Type != spec.FactTypeBoolean {
		t.Errorf("expected type boolean, got %q", fact.Type)
	}
}

func TestParseFact_Enum(t *testing.T) {
	f := writeTempFile(t, `
name: language
type: enum
values: [en, es, fr]
`)
	fact, err := spec.ParseFact(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.Type != spec.FactTypeEnum {
		t.Errorf("expected type enum, got %q", fact.Type)
	}
	if len(fact.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(fact.Values))
	}
}

func TestParseFact_ErrorOnUnknownType(t *testing.T) {
	f := writeTempFile(t, `
name: foo
type: unknown
`)
	_, err := spec.ParseFact(f)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestParseFact_ScopeDefaultsToManual(t *testing.T) {
	f := writeTempFile(t, `
name: is-logged-in
type: boolean
`)
	fact, err := spec.ParseFact(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.Scope != spec.FactScopeManual {
		t.Errorf("expected default scope manual, got %q", fact.Scope)
	}
}

func TestParseFact_ScopeParsed(t *testing.T) {
	f := writeTempFile(t, `
name: theme
type: enum
values: [light, dark]
scope: cross
`)
	fact, err := spec.ParseFact(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.Scope != spec.FactScopeCross {
		t.Errorf("expected scope cross, got %q", fact.Scope)
	}
}

func TestParseFact_ErrorOnUnknownScope(t *testing.T) {
	f := writeTempFile(t, `
name: theme
type: enum
values: [light, dark]
scope: unknown
`)
	_, err := spec.ParseFact(f)
	if err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestParseFact_ErrorOnEnumWithNoValues(t *testing.T) {
	f := writeTempFile(t, `
name: language
type: enum
`)
	_, err := spec.ParseFact(f)
	if err == nil {
		t.Fatal("expected error for enum with no values")
	}
}

func TestLoadFacts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "language.yml"), `name: language
type: enum
values: [en, es]
`)
	writeFile(t, filepath.Join(dir, "is-logged-in.yml"), `name: is-logged-in
type: boolean
`)

	facts, err := spec.LoadFacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(facts) != 2 {
		t.Errorf("expected 2 facts, got %d", len(facts))
	}
}

// helpers

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
