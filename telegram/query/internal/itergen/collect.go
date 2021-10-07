package main

import (
	"flag"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/query/internal/genutil"
)

type method struct {
	name       string
	f          *types.Func
	sig        *types.Signature
	reqType    types.Type
	resultType types.Type

	fromRequest []RequestArgument
	params      []Param
}

type collector struct {
	ignoreFields       map[string]struct{}
	canFillFromRequest map[string]struct{}
	requiredByIter     []string
	required           map[string]string

	pkg    *packages.Package
	ifaces *genutil.Interfaces

	iface          *types.Interface
	resultTypeName string
	elemName       string
	prefix         string
	pkgName        string
	requestFields  []Param
}

type collectorConfig struct {
	ResultName string
	ElemName   string
	Prefix     string
	PkgName    string
}

func (c *collectorConfig) fromFlags(set *flag.FlagSet) {
	set.StringVar(&c.ResultName, "result", "MessagesMessagesClass", "result type name")
	set.StringVar(&c.ElemName, "elem", "Elem", "element type name")
	set.StringVar(&c.Prefix, "prefix", "Messages", "prefix of methods to trim")
	set.StringVar(&c.PkgName, "package", "messages", "name of package name to generate")
}

func newCollector(pkg *packages.Package, cfg collectorConfig) *collector {
	intGetter := types.NewSignature(nil, nil,
		types.NewTuple(types.NewVar(0, nil, "", types.Typ[types.Int])), false) // func() int
	methods := []*types.Func{
		types.NewFunc(token.NoPos, nil, "GetLimit", intGetter),
	}
	match := types.NewInterfaceType(methods, nil).Complete()

	canFillFromRequest := map[string]struct{}{
		"AddOffset":  {},
		"OffsetID":   {},
		"OffsetDate": {},
		"OffsetPeer": {},
		"OffsetRate": {},
		"Offset":     {},
	}
	ignoreFields := map[string]struct{}{
		// Already handled by match interface.
		"Limit": {},
		// Not real field.
		"Flags": {},
		// Telegram ignores MaxID and MinID sometimes.
		"MaxID": {}, "MinID": {},
		// ExcludePinned used by iterator.
		"ExcludePinned": {},
		// Hash can be used internally, so do not expose it.
		"Hash": {},
	}
	requiredByIter := []string{
		"OffsetID",
		"OffsetDate",
		"Offset",
	}
	required := map[string]string{
		"Peer":    "InputPeerClass",
		"Channel": "InputChannelClass",
		"UserID":  "InputUserClass",
	}

	return &collector{
		ignoreFields:       ignoreFields,
		canFillFromRequest: canFillFromRequest,
		requiredByIter:     requiredByIter,
		required:           required,
		ifaces:             genutil.NewInterfaces(pkg),
		pkg:                pkg,
		iface:              match,
		resultTypeName:     cfg.ResultName,
		elemName:           cfg.ElemName,
		prefix:             cfg.Prefix,
		pkgName:            cfg.PkgName,
	}
}

func (c *collector) methods() ([]method, error) { // nolint:gocognit
	var result []method

	for _, def := range genutil.Funcs(c.pkg, func(f genutil.Func) bool {
		return f.Args().Len() == 2 && f.Results().Len() == 2
	}) {
		args := def.Args()
		results := def.Results()

		ptr, ok := args.At(1).Type().(*types.Pointer)
		if !ok || !types.Implements(ptr, c.iface) {
			continue
		}
		reqType := ptr.Elem()

		resultType, ok := results.At(0).Type().(*types.Named)
		if !ok {
			continue
		}

		if resultType.Obj().Name() != c.resultTypeName {
			continue
		}
		name := strings.TrimPrefix(def.Decl.Name(), c.prefix)

		m := method{
			name:       name,
			f:          def.Decl,
			sig:        def.Sig,
			reqType:    reqType,
			resultType: resultType,
		}

		reqTypeStruct, ok := reqType.Underlying().(*types.Struct)
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", reqType.Underlying())
		}

		for i := 0; i < reqTypeStruct.NumFields(); i++ {
			field := reqTypeStruct.Field(i)

			if _, ok := c.ignoreFields[field.Name()]; ok {
				continue
			}

			param := varToParam(field)
			if _, ok := c.canFillFromRequest[field.Name()]; ok {
				requiredByIter := false
				for _, field := range c.requiredByIter {
					if field == param.OriginalName {
						requiredByIter = true
						break
					}
				}
				m.fromRequest = append(m.fromRequest, RequestArgument{
					Arg:            param,
					Chain:          field.Name() == "OffsetID" || field.Name() == "OffsetDate",
					RequiredByIter: requiredByIter,
				})

				skip := false
				for _, field := range c.requestFields {
					if field.OriginalName == param.OriginalName {
						skip = true
						break
					}
				}
				if !skip {
					c.requestFields = append(c.requestFields, param)
				}
				continue
			}

			m.params = append(m.params, param)
		}

		result = append(result, m)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].name < result[j].name
	})
	return result, nil
}

func (c *collector) Config() (Config, error) {
	methods, err := c.collect()
	if err != nil {
		return Config{}, xerrors.Errorf("collect: %w", err)
	}

	return Config{
		Methods:       methods,
		Package:       c.pkgName,
		ResultName:    c.resultTypeName,
		RequestFields: sortParams(c.requestFields),
	}, nil
}

func (c *collector) collect() ([]Method, error) {
	methods, err := c.methods()
	if err != nil {
		return nil, xerrors.Errorf("collect types: %w", err)
	}

	result := make([]Method, 0, len(methods))
	for _, method := range methods {
		mapping := method.fromRequest
		sort.SliceStable(mapping, func(i, j int) bool {
			return mapping[i].Arg.Name < mapping[j].Arg.Name
		})

		m := Method{
			Name:              method.name,
			OriginalName:      method.f.Name(),
			RequestName:       genutil.PrintType(method.reqType),
			ResultName:        genutil.PrintType(method.resultType),
			AdditionalMapping: mapping,
			AdditionalParams:  sortParams(method.params),
			IteratorName:      "Iterator",
			ElemName:          c.elemName,
		}

		for _, field := range method.params {
			if _, ok := c.required[field.OriginalName]; ok {
				m.RequiredParams = append(m.RequiredParams, field)
			}
		}
		m.RequiredParams = sortParams(m.RequiredParams)

		cases, err := c.collectSpecial(m)
		if err != nil {
			return nil, xerrors.Errorf("collect special: %w", err)
		}

		m.SpecialCase = cases
		result = append(result, m)
	}
	return result, nil
}
