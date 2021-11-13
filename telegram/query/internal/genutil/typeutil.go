package genutil

import (
	"go/types"

	"github.com/go-faster/errors"
	"golang.org/x/tools/go/packages"
)

// Implementations finds iface implementations.
func Implementations(pkg *packages.Package, iface *types.Interface) []*types.Named {
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

// Interfaces is a simple utility struct to find interfaces and implementations.
type Interfaces struct {
	pkg        *packages.Package
	implsCache map[string][]*types.Named
}

// NewInterfaces creates new Interfaces structure.
func NewInterfaces(pkg *packages.Package) *Interfaces {
	return &Interfaces{pkg: pkg, implsCache: map[string][]*types.Named{}}
}

// Interface finds interface by name.
func (c *Interfaces) Interface(name string) (*types.Interface, error) {
	obj := c.pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, errors.Errorf("%q not found", name)
	}

	v, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, errors.Errorf("%q has unexpected kind type %T", name, obj.Type().Underlying())
	}

	return v, nil
}

// Implementations finds interface implementations by interface name.
func (c *Interfaces) Implementations(name string) ([]*types.Named, error) {
	impls, ok := c.implsCache[name]
	if ok {
		return impls, nil
	}

	iface, err := c.Interface(name)
	if err != nil {
		return nil, errors.Wrapf(err, "find %q", name)
	}

	impls = Implementations(c.pkg, iface)
	c.implsCache[name] = impls
	return impls, nil
}
