package genutil

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/ogen-go/errors"
	"golang.org/x/tools/go/packages"
)

func loadPackages(ctx context.Context, dir, pattern string, environ []string) ([]*packages.Package, error) {
	return packages.Load(&packages.Config{
		Context: ctx,
		Dir:     dir,
		Mode: packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedDeps,
		Env:  environ,
		Fset: token.NewFileSet(),
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			const mode = parser.AllErrors | parser.ParseComments
			return parser.ParseFile(fset, filename, src, mode)
		},
	}, pattern)
}

// Load loads package using given pattern.
func Load(ctx context.Context, pattern string) (*packages.Package, error) {
	pkgs, err := loadPackages(ctx, "", pattern, os.Environ())
	if err != nil {
		return nil, errors.Wrap(err, "load packages")
	}

	for _, pkg := range pkgs {
		if pkg.ID == pattern {
			return pkg, nil
		}
	}

	return nil, errors.Errorf("package %s not found", pattern)
}
