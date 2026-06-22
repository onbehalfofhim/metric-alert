/*
Package noosexit содержит анализатор,
запрещающий прямой вызов os.Exit внутри функции main пакета main.
*/
package exitcheck

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "запрещает прямой вызов os.Exit внутри main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" ||
		strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil

	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				obj := pass.TypesInfo.Uses[pkgIdent]
				if obj == nil {
					return true
				}

				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				if pkgName.Imported().Path() == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(
						call.Pos(),
						"direct call to os.Exit in main function is forbidden",
					)
				}
				return true
			})
		}
	}

	return nil, nil
}
