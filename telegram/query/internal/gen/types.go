package main

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
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

func load(ctx context.Context, pattern string) (*packages.Package, error) {
	pkgs, err := loadPackages(ctx, "", pattern, os.Environ())
	if err != nil {
		return nil, xerrors.Errorf("load packages: %w", err)
	}

	for _, pkg := range pkgs {
		if pkg.ID == pattern {
			return pkg, nil
		}
	}

	return nil, xerrors.Errorf("package %s not found", pattern)
}

func printType(typ types.Type) string {
	return types.TypeString(typ, func(i *types.Package) string {
		return i.Name()
	})
}

func (c *collector) findInterface(pkg *packages.Package, name string) (*types.Interface, error) {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, xerrors.Errorf("%q not found", name)
	}

	v, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, xerrors.Errorf("%q has unexpected kind type %T", name, obj.Type().Underlying())
	}

	return v, nil
}

func (c *collector) collectImplementations(pkg *packages.Package, iface *types.Interface) []*types.Named {
	var r []*types.Named

	for _, def := range pkg.TypesInfo.Defs {
		if def == nil || !def.Exported() {
			continue
		}

		named, ok := def.Type().(*types.Named)
		if !ok {
			continue
		}

		if !types.Implements(types.NewPointer(named), iface) {
			continue
		}

		r = append(r, named)
	}

	return r
}

func (c *collector) findImplementations(pkg *packages.Package, name string) ([]*types.Named, error) {
	impls, ok := c.implsCache[name]
	if ok {
		return impls, nil
	}

	iface, err := c.findInterface(pkg, name)
	if err != nil {
		return nil, xerrors.Errorf("find %q: %w", name, err)
	}

	impls = c.collectImplementations(pkg, iface)
	c.implsCache[name] = impls
	return impls, nil
}

func varToParam(field *types.Var) Param {
	fieldName := field.Name()
	fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
	return Param{
		Name:         fieldName,
		OriginalName: field.Name(),
		Type:         printType(field.Type()),
	}
}
