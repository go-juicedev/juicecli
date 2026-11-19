package internal

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/go-juicedev/juice"
	"github.com/go-juicedev/juice/driver"
	"github.com/go-juicedev/juice/eval"
	juicesql "github.com/go-juicedev/juice/sql"
	astlite "github.com/go-juicedev/juicecli/internal/ast"
)

func TestFormatParamsUsesSyntheticNameWhenParamIsUnnamed(t *testing.T) {
	src := `package repo

import "context"

type User struct{}

type Repo interface {
	Get(context.Context, int64) (User, error)
	List(context.Context, []int64) ([]User, error)
	Create(context.Context, *User) error
	Delete(context.Context, int64, string) (error)
	Named(ctx context.Context, id int64) (User, error)
}
`
	methods := interfaceMethods(t, src)

	tests := []struct {
		name string
		want string
	}{
		{name: "Get", want: `juice.H{"arg1": arg1}`},
		{name: "List", want: `juice.H{"arg1": arg1}`},
		{name: "Create", want: "arg1"},
		{name: "Delete", want: `juice.H{"arg1": arg1, "arg2": arg2}`},
		{name: "Named", want: `juice.H{"id": id}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatParams(methods[tt.name].Params())
			if got != tt.want {
				t.Fatalf("formatParams() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnnamedParamsGenerateCompilableMethod(t *testing.T) {
	src := `package repo

import "context"

type User struct{}

type Repo interface {
	Get(context.Context, int64) (User, error)
	Create(context.Context, *User) error
	Delete(context.Context, int64, string) error
}
`
	methods := interfaceMethods(t, src)
	var b strings.Builder
	b.WriteString("package repo\n")
	for _, name := range []string{"Get", "Create", "Delete"} {
		method := methods[name]
		function := &Function{
			method:   method,
			receiver: "RepoImpl",
			typename: "Repo",
		}
		action := juicesql.Select
		if name != "Get" {
			action = juicesql.Insert
		}
		maker := &GenericFunctionBodyMaker{
			statement: statementStub{action: action},
			function:  function,
			readFuncBodyMakerProvider: func(statement juice.Statement, function *Function) FunctionBodyMaker {
				return &readFuncBodyMakerV2{readFuncBodyMaker: &readFuncBodyMaker{statement: statement, function: function}}
			},
			writeFuncBodyMakerProvider: func(statement juice.Statement, function *Function) FunctionBodyMaker {
				return &writeFuncBodyMakerV2{writeFuncBodyMaker: &writeFuncBodyMaker{statement: statement, function: function}}
			},
		}
		if err := maker.Make(); err != nil {
			t.Fatalf("Make(%s) error = %v", name, err)
		}
		b.WriteString(function.String())
		b.WriteString("\n\n")
	}

	generated := b.String()
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", generated, parser.SkipObjectResolution); err != nil {
		t.Fatalf("generated code does not parse: %v\n%s", err, generated)
	}
	for _, want := range []string{
		"func (r RepoImpl) Get(arg0 context.Context, arg1 int64)",
		`juice.H{"arg1": arg1}`,
		"func (r RepoImpl) Create(arg0 context.Context, arg1 *User)",
		"juice.ExecContext(arg0, r.manager, Repo(r).Create, arg1)",
		`juice.H{"arg1": arg1, "arg2": arg2}`,
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated code missing %q\n%s", want, generated)
		}
	}
}

func interfaceMethods(t *testing.T, src string) map[string]*astlite.Function {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "iface.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	methods := map[string]*astlite.Function{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*goast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*goast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Repo" {
				continue
			}
			iface, ok := typeSpec.Type.(*goast.InterfaceType)
			if !ok {
				t.Fatal("Repo is not an interface")
			}
			for _, method := range (&astlite.Interface{InterfaceType: iface}).Methods() {
				methods[method.Name()] = method
			}
		}
	}
	if len(methods) == 0 {
		t.Fatal("no methods parsed")
	}
	return methods
}

type statementStub struct {
	action juicesql.Action
}

func (statementStub) ID() juice.StatementID { return "repo.Repo.Method" }
func (statementStub) Attribute(string) string { return "" }
func (s statementStub) Action() juicesql.Action { return s.action }
func (statementStub) ResultMap() (juicesql.ResultMap, error) {
	return nil, juicesql.ErrResultMapNotSet
}
func (statementStub) Build(driver.Translator, eval.Parameter) (string, []any, error) {
	return "select 1", nil, nil
}
