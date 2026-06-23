package main

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestHasGenerateResetComment(t *testing.T) {
	tests := []struct {
		name string
		doc  *ast.CommentGroup
		want bool
	}{
		{
			name: "nil doc",
			doc:  nil,
			want: false,
		},
		{
			name: "contains marker",
			doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// generate:reset"},
				},
			},
			want: true,
		},
		{
			name: "other comment",
			doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// some comment"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasGenerateResetComment(tt.doc))
		})
	}
}

func TestZeroValue(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		want string
	}{
		{
			name: "string",
			typ:  types.Typ[types.String],
			want: `""`,
		},
		{
			name: "bool",
			typ:  types.Typ[types.Bool],
			want: "false",
		},
		{
			name: "int",
			typ:  types.Typ[types.Int],
			want: "0",
		},
		{
			name: "slice",
			typ:  types.NewSlice(types.Typ[types.String]),
			want: "nil",
		},
		{
			name: "map",
			typ: types.NewMap(
				types.Typ[types.String],
				types.Typ[types.Int],
			),
			want: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, zeroValue(tt.typ))
		})
	}
}

func TestHasResetMethod(t *testing.T) {
	src := `
		package test

		type Child struct{}

		func (c *Child) Reset() {}

		type Other struct{}
	`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{
		Importer: importer.Default(),
	}

	pkg, err := conf.Check(
		"test",
		fset,
		[]*ast.File{file},
		info,
	)
	require.NoError(t, err)

	childObj := pkg.Scope().Lookup("Child")
	otherObj := pkg.Scope().Lookup("Other")

	assert.True(t, hasResetMethod(childObj.Type()))
	assert.False(t, hasResetMethod(otherObj.Type()))
}

func TestGenerateReset(t *testing.T) {
	src := `
		package test

		type Child struct{}

		func (c *Child) Reset() {}

		// generate:reset
		type Metric struct {
			ID string
			Tags []string
			Data map[string]int
			Child Child
			Count *int
		}
	`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{
		Importer: importer.Default(),
	}

	_, err = conf.Check(
		"test",
		fset,
		[]*ast.File{file},
		info,
	)
	require.NoError(t, err)

	var st *ast.StructType

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			ts := spec.(*ast.TypeSpec)

			if ts.Name.Name == "Metric" {
				st = ts.Type.(*ast.StructType)
			}
		}
	}

	require.NotNil(t, st)

	var out bytes.Buffer

	generateReset(&out, "Metric", st, info)

	got := out.String()

	assert.Contains(t, got, `s.ID = ""`)
	assert.Contains(t, got, `s.Tags = s.Tags[:0]`)
	assert.Contains(t, got, `clear(s.Data)`)
	assert.Contains(t, got, `s.Child.Reset()`)
	assert.Contains(t, got, `*s.Count = 0`)
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()

	pkg := &packages.Package{
		Name: "testpkg",
		GoFiles: []string{
			filepath.Join(dir, "main.go"),
		},
	}

	writeFile(pkg, []byte(`func Test() {}`))

	path := filepath.Join(dir, "reset.gen.go")

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)

	assert.True(t, strings.Contains(content, "package testpkg"))
	assert.True(t, strings.Contains(content, "func Test()"))
}
