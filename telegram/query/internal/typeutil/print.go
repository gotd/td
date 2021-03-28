package typeutil

import "go/types"

// PrintType prints typename into string without package name.
func PrintType(typ types.Type) string {
	return types.TypeString(typ, func(i *types.Package) string {
		return i.Name()
	})
}
