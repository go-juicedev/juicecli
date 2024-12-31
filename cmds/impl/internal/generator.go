package internal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-juicedev/juice"
)

type Generator struct {
	cfg       juice.IConfiguration
	impl      *Implement
	namespace string
}

func (g *Generator) Generate() (io.Reader, error) {
	for _, m := range g.impl.iface.Methods() {
		method := m
		key := fmt.Sprintf("%s.%s", g.namespace, method.Name())
		statement, err := g.cfg.GetStatement(key)
		if err != nil {
			return nil, err
		}
		if statement.Attribute("gen") == "false" || statement.Attribute("generate") == "false" { // skip
			continue
		}
		function := &Function{method: method, receiver: g.impl.dst, typename: g.impl.src}
		maker := FunctionBodyMaker{statement: statement, function: function}
		if err = maker.Make(); err != nil {
			return nil, err
		}
		g.impl.methods = append(g.impl.methods, function)
	}
	builder := strings.Builder{}
	args := strings.Join(os.Args[:], " ")
	builder.WriteString(fmt.Sprintf("// Code generated by \"%s\"; DO NOT EDIT.", args))
	builder.WriteString("\n\n")
	builder.WriteString(g.impl.String())
	return strings.NewReader(builder.String()), nil
}

// WriteTo writes generated code to writer.
func (g *Generator) WriteTo(writer io.Writer) (int64, error) {
	reader, err := g.Generate()
	if err != nil {
		return 0, err
	}
	return io.Copy(writer, reader)
}

func NewGenerator(namespace string, cfg juice.IConfiguration, impl *Implement) *Generator {
	return &Generator{cfg: cfg, impl: impl, namespace: namespace}
}