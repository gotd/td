package main

import (
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/nnqq/td/telegram/query/internal/genutil"
)

func isHashField(field *types.Var) bool {
	basic, ok := field.Type().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Kind() == types.Int64 && field.Name() == "Hash"
}

func hasHashField(st *types.Struct) bool {
	for i := 0; i < st.NumFields(); i++ {
		if isHashField(st.Field(i)) {
			return true
		}
	}

	return false
}

type request struct {
	name   string
	params []Param
}

func isCachedQuery(args *types.Tuple) (request, bool) {
	arg := args.At(1)
	switch req := arg.Type().(type) {
	case *types.Pointer:
		named, ok := req.Elem().(*types.Named)
		if !ok {
			return request{}, false
		}

		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			return request{}, false
		}

		var r []Param
		for i := 0; i < st.NumFields(); i++ {
			field := st.Field(i)
			if strings.Contains(field.Name(), "Offset") {
				return request{}, false
			}

			if isHashField(field) || field.Name() == "Flags" {
				continue
			}

			r = append(r, varToParam(field))
		}

		return request{
			name:   named.Obj().Name(),
			params: sortParams(r),
		}, hasHashField(st)
	case *types.Basic:
		if req.Kind() != types.Int64 || arg.Name() != "hash" {
			return request{}, false
		}
		return request{}, true
	default:
		return request{}, false
	}
}

func collect(pkg *packages.Package) []CachedQuery {
	var r []CachedQuery

	for _, def := range genutil.Funcs(pkg, func(f genutil.Func) bool {
		return f.Args().Len() == 2 && f.Results().Len() == 2
	}) {
		args := def.Args()
		req, ok := isCachedQuery(args)
		if !ok {
			continue
		}

		resultNamed, ok := def.Results().At(0).Type().(*types.Named)
		if !ok {
			continue
		}

		result, ok := resultNamed.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		impls := genutil.Implementations(pkg, result)
		if len(impls) != 2 {
			continue
		}
		var (
			notModified *types.Named
			pure        *types.Named
		)
		for _, impl := range impls {
			if notModified == nil && strings.Contains(impl.Obj().Name(), "NotModified") {
				notModified = impl
				continue
			}

			if pure == nil {
				pure = impl
			}
		}
		if pure == nil || notModified == nil {
			continue
		}

		pureStruct, ok := pure.Underlying().(*types.Struct)
		if !ok {
			continue
		}

		r = append(r, CachedQuery{
			Name:            def.Decl.Name(),
			MethodName:      def.Decl.Name(),
			RequestName:     req.name,
			ManualHash:      !hasHashField(pureStruct),
			RequestParams:   req.params,
			ResultName:      pure.Obj().Name(),
			NotModifiedName: notModified.Obj().Name(),
		})
	}
	sort.SliceStable(r, func(i, j int) bool {
		return r[i].Name < r[j].Name
	})

	return r
}
