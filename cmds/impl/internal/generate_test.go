package internal

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-juicedev/juice"
	configparser "github.com/go-juicedev/juice/parser"
	"github.com/go-juicedev/juice/parser/xml"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerateInterfaceImplGolden(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata")
	namespace := "github.com.go-juicedev.juicecli.cmds.impl.testdata.Interface"

	iface, file := parseInterface(t, filepath.Join(testdataDir, "interface.go"), "Interface")
	catalog, mappers := loadTestMappers(t, filepath.Join(testdataDir, "config", "juice.xml"))

	implement, err := NewImplement(file, iface, catalog, mappers, namespace, "Interface", "InterfaceImpl")
	if err != nil {
		t.Fatalf("NewImplement() error = %v", err)
	}

	reader, err := NewGenerator(implement).Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read generated code: %v", err)
	}

	goldenPath := filepath.Join(testdataDir, "interface_impl.go")
	if *updateGolden {
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated code mismatch\n=== got ===\n%s\n=== want ===\n%s", got, want)
	}
}

func TestGenerateRejectsExtraMapperStatement(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata")
	namespace := "github.com.go-juicedev.juicecli.cmds.impl.testdata.Interface"

	iface, file := parseInterface(t, filepath.Join(testdataDir, "interface.go"), "Interface")
	catalog, mappers := loadTestMappers(t, filepath.Join(testdataDir, "config", "juice.xml"))

	for i := range mappers {
		if mappers[i].Namespace != namespace {
			continue
		}
		orphan := mappers[i].Statements[0]
		orphan.ID = "OrphanMethod"
		mappers[i].Statements = append(mappers[i].Statements, orphan)
		break
	}

	implement, err := NewImplement(file, iface, catalog, mappers, namespace, "Interface", "InterfaceImpl")
	if err != nil {
		t.Fatalf("NewImplement() error = %v", err)
	}
	_, err = NewGenerator(implement).Generate()
	if err == nil {
		t.Fatal("Generate() error = nil, want extra statement coverage error")
	}
	if !strings.Contains(err.Error(), "extra in mapper: OrphanMethod") {
		t.Fatalf("Generate() error = %v, want extra OrphanMethod", err)
	}
}

func parseInterface(t *testing.T, path, typeName string) (*ast.InterfaceType, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != typeName {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				t.Fatalf("%s is not an interface", typeName)
			}
			return iface, file
		}
	}
	t.Fatalf("type %s not found in %s", typeName, path)
	return nil, nil
}

func loadTestMappers(t *testing.T, configPath string) (juice.StatementCatalog, []configparser.Mapper) {
	t.Helper()
	dirname := filepath.Dir(configPath)
	filename := filepath.Base(configPath)

	root, err := os.OpenRoot(dirname)
	if err != nil {
		t.Fatalf("OpenRoot(%s): %v", dirname, err)
	}
	t.Cleanup(func() { _ = root.Close() })

	reader, err := root.Open(filename)
	if err != nil {
		t.Fatalf("open config: %v", err)
	}
	defer func() { _ = reader.Close() }()

	mappers, err := xml.ParseMappers(root.FS(), reader)
	if err != nil {
		t.Fatalf("ParseMappers: %v", err)
	}
	catalog, err := juice.CompileMappers(mappers)
	if err != nil {
		t.Fatalf("CompileMappers: %v", err)
	}
	return catalog, mappers
}
