package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/go-juicedev/juice"
	configparser "github.com/go-juicedev/juice/parser"
	astlite "github.com/go-juicedev/juicecli/internal/ast"
)

type Implement interface {
	Render() (string, error)
}

type implement struct {
	iface                     *astlite.Interface
	file                      *ast.File
	statementCatalog          juice.StatementCatalog
	mappers                   []configparser.Mapper
	extraImports              astlite.ImportGroup
	methods                   FunctionGroup
	src, dst                  string
	namespace                 string
	functionBodyMakerProvider FunctionBodyMakerProvider
}

func (i *implement) Package() string {
	return i.file.Name.Name
}

func (i *implement) Imports() astlite.ImportGroup {
	return append(i.iface.Imports(i.file.Imports), i.extraImports...).Uniq()
}

func (i *implement) buildFunction() error {
	methods := i.iface.Methods()
	methodNames := make([]string, 0, len(methods))
	for _, method := range methods {
		methodNames = append(methodNames, method.Name())
	}
	if err := checkStatementCoverage(i.namespace, methodNames, i.mappers); err != nil {
		return err
	}

	for _, m := range methods {
		method := m
		key := fmt.Sprintf("%s.%s", i.namespace, method.Name())
		statement, err := i.statementCatalog.Statement(juice.StatementID(key))
		if err != nil {
			// checkStatementCoverage should have already reported missing ids.
			return err
		}
		if statement.Attribute("gen") == "false" || statement.Attribute("generate") == "false" { // skip
			continue
		}
		function := &Function{
			method:      method,
			receiver:    i.dst,
			typename:    i.src,
			statementID: statement.ID().String(),
		}
		maker := i.functionBodyMakerProvider(statement, function)
		if err = maker.Make(); err != nil {
			return err
		}
		i.methods = append(i.methods, function)
	}

	return nil
}

func NewImplement(writer *ast.File, iface *ast.InterfaceType, statementCatalog juice.StatementCatalog, mappers []configparser.Mapper, namespace, input, output string) (Implement, error) {
	impl := &implement{
		dst:              output,
		statementCatalog: statementCatalog,
		mappers:          mappers,
		src:              input,
		file:             writer,
		iface:            &astlite.Interface{InterfaceType: iface},
		extraImports: astlite.ImportGroup{
			&astlite.Import{ImportSpec: extraImport.Imports[0]},
		},
		namespace: namespace,
	}

	impl.functionBodyMakerProvider = func(statement juice.Statement, function *Function) FunctionBodyMaker {
		return &GenericFunctionBodyMaker{
			statement: statement,
			function:  function,
			readFuncBodyMakerProvider: func(statement juice.Statement, function *Function) FunctionBodyMaker {
				return &readFuncBodyMakerV2{
					readFuncBodyMaker: &readFuncBodyMaker{
						statement: statement,
						function:  function,
					}}
			},
			writeFuncBodyMakerProvider: func(statement juice.Statement, function *Function) FunctionBodyMaker {
				return &writeFuncBodyMakerV2{
					writeFuncBodyMaker: &writeFuncBodyMaker{
						statement: statement,
						function:  function,
					},
				}
			},
		}
	}
	return &ImplementV2{implement: impl}, nil
}

type ImplementV2 struct {
	*implement
}

func (i *ImplementV2) constructor() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("// New%s returns a new %s.\n", i.src, i.src))
	builder.WriteString(fmt.Sprintf("func New%s(manager juice.Manager) %s {", i.src, i.src))
	builder.WriteString("\n\t")
	builder.WriteString(fmt.Sprintf("return &%s{manager: manager}", i.dst))
	builder.WriteString("\n")
	builder.WriteString("}")
	return builder.String()
}

func (i *ImplementV2) Render() (string, error) {
	if err := i.implement.buildFunction(); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("package %s", i.Package()))
	builder.WriteString("\n\n")
	builder.WriteString(i.Imports().String())
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("type %s struct { manager juice.Manager }", i.dst))
	builder.WriteString("\n\n")
	// implement methods
	builder.WriteString(i.methods.String())
	builder.WriteString("\n\n")
	builder.WriteString(i.constructor())
	return formatCode(builder.String()), nil
}

var extraImportSrc = `
package main

import "github.com/go-juicedev/juice"
`

// extraImport is an ast.File for extra import.
var extraImport *ast.File

func init() {
	var err error
	extraImport, err = parser.ParseFile(token.NewFileSet(), "", extraImportSrc, parser.ImportsOnly)
	if err != nil {
		log.Fatal(err)
	}
}
