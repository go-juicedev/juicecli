package internal

import "testing"

func TestParserNamespaceUsesExplicitValue(t *testing.T) {
	const want = "custom.repo.UserRepository"
	got, err := NewParser("UserRepository").WithNamespace(want).Namespace()
	if err != nil {
		t.Fatalf("Namespace() error = %v", err)
	}
	if got != want {
		t.Fatalf("Namespace() = %q, want %q", got, want)
	}
}

func TestParserNamespaceFallsBackToAutocomplete(t *testing.T) {
	got, err := NewParser("Interface").Namespace()
	if err != nil {
		t.Fatalf("Namespace() error = %v", err)
	}
	if got == "" {
		t.Fatal("Namespace() returned empty string")
	}
	// Autocomplete should still end with the type name.
	if wantSuffix := ".Interface"; len(got) < len(wantSuffix) || got[len(got)-len(wantSuffix):] != wantSuffix {
		t.Fatalf("Namespace() = %q, want suffix %q", got, wantSuffix)
	}
}
