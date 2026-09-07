package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-juicedev/juice"
	"github.com/go-juicedev/juice/parser/xml"
	"github.com/go-juicedev/juicecli/internal/module"
	"github.com/go-juicedev/juicecli/internal/namespace"
)

// defaultConfigFiles is the default config file name
// while config is not set, we will check if config.xml or config/config.xml exists
var defaultConfigFiles = [...]string{
	"juice.xml",
	"config.xml",
	"config/juice.xml",
	"config/config.xml",
}

func NewParser(typeName string) *Parser {
	return &Parser{typename: typeName}
}

type Parser struct {
	typename  string
	impl      string
	cfg       string
	namespace string
	output    string
}

func (p *Parser) WithConfig(cfg string) *Parser {
	p.cfg = cfg
	return p
}

func (p *Parser) WithNamespace(namespace string) *Parser {
	p.namespace = namespace
	return p
}

func (p *Parser) WithOutput(output string) *Parser {
	p.output = output
	return p
}

func (p *Parser) WithImpl(impl string) *Parser {
	p.impl = impl
	return p
}

func (p *Parser) config() (string, error) {
	if p.cfg != "" {
		return p.cfg, nil
	}
	for _, defaultConfigFile := range defaultConfigFiles {
		exists, err := fileExists(defaultConfigFile)
		if err != nil {
			return "", err
		}
		if exists {
			return defaultConfigFile, nil
		}
	}
	return "", errors.New(strings.Join(defaultConfigFiles[:], "|") + " not found")
}

func (p *Parser) Config() (juice.StatementCatalog, error) {
	config, err := p.config()
	if err != nil {
		return nil, err
	}
	filename := config
	dirname := filepath.Dir(filename)
	filename = filepath.Base(filename)

	root, err := os.OpenRoot(dirname)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	reader, err := root.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	mappers, err := xml.ParseMappers(root.FS(), reader)
	if err != nil {
		return nil, err
	}

	return juice.CompileMappers(mappers)
}

func (p *Parser) TypeInterface() (*ast.InterfaceType, *ast.File, error) {
	node, file, err := module.FindTypeNode("./", p.typename)
	if err != nil {
		return nil, nil, fmt.Errorf("can not find type %s", p.typename)
	}
	iface, ok := node.(*ast.InterfaceType)
	if !ok {
		return nil, nil, fmt.Errorf("%s is not an interface", p.typename)
	}
	return iface, file, nil
}

func (p *Parser) Output() (io.Writer, error) {
	if p.output == "" {
		return os.Stdout, nil
	}
	return os.Create(p.output)
}

func (p *Parser) Namespace() (string, error) {
	if p.namespace != "" {
		return p.namespace, nil
	}
	cmp := namespace.AutoComplete{TypeName: p.typename}
	return cmp.Autocomplete()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
