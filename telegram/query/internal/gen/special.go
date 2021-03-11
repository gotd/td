package main

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"
)

func (c *collector) collectSpecial(pkg *packages.Package, m Method) ([]SpecialCaseChain, error) {
	var r []SpecialCaseChain
	for _, field := range m.AdditionalParams {
		if field.Type == "tg.MessagesFilterClass" {
			iface, err := findInterface(pkg, "MessagesFilterClass")
			if err != nil {
				return nil, xerrors.Errorf("find MessagesFilterClass: %w", err)
			}

			impls := collectImplementations(pkg, iface)
			for _, impl := range impls {
				s, ok := impl.Underlying().(*types.Struct)
				if !ok {
					continue
				}

				cse := SpecialCaseChain{
					ConstructorName: strings.TrimPrefix(impl.Obj().Name(), "InputMessagesFilter"),
					ConstructorType: printType(impl),
					Field:           field,
				}

				if strings.Contains(cse.ConstructorName, "Empty") {
					continue
				}

				for i := 0; i < s.NumFields(); i++ {
					field := s.Field(i)
					if field.Name() == "Flags" {
						continue
					}

					cse.Args = append(cse.Args, varToParam(field))
				}

				r = append(r, cse)
			}
		}
	}
	return r, nil
}
