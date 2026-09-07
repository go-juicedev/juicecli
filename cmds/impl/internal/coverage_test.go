package internal

import (
	"errors"
	"strings"
	"testing"

	configparser "github.com/go-juicedev/juice/parser"
)

func TestCheckStatementCoverageOK(t *testing.T) {
	err := checkStatementCoverage("demo.Repo", []string{"Find", "Create"}, []configparser.Mapper{
		{
			Namespace: "demo.Repo",
			Statements: []configparser.Statement{
				{ID: "Find"},
				{ID: "Create"},
			},
		},
		// Other namespaces are ignored.
		{
			Namespace:  "demo.Other",
			Statements: []configparser.Statement{{ID: "OnlyHere"}},
		},
	})
	if err != nil {
		t.Fatalf("checkStatementCoverage() error = %v", err)
	}
}

func TestCheckStatementCoverageMissingAndExtra(t *testing.T) {
	err := checkStatementCoverage("demo.Repo", []string{"Find", "Missing"}, []configparser.Mapper{
		{
			Namespace: "demo.Repo",
			Statements: []configparser.Statement{
				{ID: "Find"},
				{ID: "Extra"},
			},
		},
	})
	if err == nil {
		t.Fatal("checkStatementCoverage() error = nil, want mismatch")
	}
	var coverageErr *statementCoverageError
	if !errors.As(err, &coverageErr) {
		t.Fatalf("error type = %T, want *statementCoverageError", err)
	}
	msg := err.Error()
	for _, want := range []string{
		`namespace "demo.Repo"`,
		"missing in mapper: Missing",
		"extra in mapper: Extra",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
}
