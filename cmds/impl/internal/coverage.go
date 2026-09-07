package internal

import (
	"fmt"
	"sort"
	"strings"

	configparser "github.com/go-juicedev/juice/parser"
)

// statementCoverageError reports mapper statements that do not line up with
// interface methods under the same namespace.
type statementCoverageError struct {
	namespace string
	missing   []string // interface methods with no mapper statement
	extra     []string // mapper statements with no interface method
}

func (e *statementCoverageError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "statement coverage mismatch for namespace %q", e.namespace)
	if len(e.missing) > 0 {
		fmt.Fprintf(&b, "\n  missing in mapper: %s", strings.Join(e.missing, ", "))
	}
	if len(e.extra) > 0 {
		fmt.Fprintf(&b, "\n  extra in mapper: %s", strings.Join(e.extra, ", "))
	}
	return b.String()
}

// checkStatementCoverage ensures every interface method has a mapper statement
// in namespace, and every mapper statement in that namespace has an interface method.
func checkStatementCoverage(namespace string, methodNames []string, mappers []configparser.Mapper) error {
	methods := make(map[string]struct{}, len(methodNames))
	for _, name := range methodNames {
		if name == "" {
			continue
		}
		methods[name] = struct{}{}
	}

	statements := make(map[string]struct{})
	for _, mapper := range mappers {
		if mapper.Namespace != namespace {
			continue
		}
		for _, statement := range mapper.Statements {
			if statement.ID == "" {
				continue
			}
			statements[statement.ID] = struct{}{}
		}
	}

	var missing []string
	for name := range methods {
		if _, ok := statements[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)

	var extra []string
	for id := range statements {
		if _, ok := methods[id]; !ok {
			extra = append(extra, id)
		}
	}
	sort.Strings(extra)

	if len(missing) == 0 && len(extra) == 0 {
		return nil
	}
	return &statementCoverageError{
		namespace: namespace,
		missing:   missing,
		extra:     extra,
	}
}
