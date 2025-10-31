package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
	`unicode`

	astlite "github.com/go-juicedev/juicecli/internal/ast"
)

type Implement struct {
	iface        *astlite.Interface
	file         *ast.File
	extraImports astlite.ImportGroup
	methods      FunctionGroup
	src, dst     string
	version      string
}

func (i *Implement) Package() string {
	return i.file.Name.Name
}

func (i *Implement) Imports() astlite.ImportGroup {
	return append(i.iface.Imports(i.file.Imports), i.extraImports...).Uniq()
}

func (i *Implement) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("package %s", i.Package()))
	builder.WriteString("\n\n")
	builder.WriteString(i.Imports().String())
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("type %s struct { manager juice.Manager }", i.dst))
	builder.WriteString("\n\n")
	// implement methods
	builder.WriteString(fmt.Sprintf("var %s %s", lowercasing(i.dst), i.src))
	builder.WriteString("\n\n")
	builder.WriteString(i.methods.String())
	builder.WriteString("\n\n")
	builder.WriteString(i.constructor())
	return formatCode(builder.String())
}

// constructor returns a constructor for the implement.
func (i *Implement) constructor() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("// New%s returns a new %s.\n", i.src, i.src))
	switch i.version {
	case "v1":
		builder.WriteString(fmt.Sprintf("func New%s() %s {", i.src, i.src))
		builder.WriteString("\n\t")
		builder.WriteString(fmt.Sprintf("return &%s{}", i.dst))
	case "v2":
		builder.WriteString(fmt.Sprintf("func New%s(manager juice.Manager) %s {", i.src, i.src))
		builder.WriteString("\n\t")
		builder.WriteString(fmt.Sprintf("return &%s{manager: manager}", i.dst))
	}
	builder.WriteString("\n")
	builder.WriteString("}")
	return builder.String()
}

func NewImplement(writer *ast.File, iface *ast.InterfaceType, version, input, output string) *Implement {
	impl := &Implement{
		dst:   output,
		src:   input,
		file:  writer,
		iface: &astlite.Interface{InterfaceType: iface},
		extraImports: astlite.ImportGroup{
			&astlite.Import{ImportSpec: extraImport.Imports[0]},
		},
		version: version,
	}
	return impl
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

func lowercasing(text string) string {
	if text == "" {
		return text
	}
	return string(unicode.ToLower(rune(text[0]))) + text[1:]
}
