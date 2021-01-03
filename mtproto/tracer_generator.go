// +build tools

package main

import (
	"fmt"
	"go/ast"
	"strings"

	builders "github.com/tdakkota/astbuilders"
	macro "github.com/tdakkota/gomacro"
	"github.com/tdakkota/gomacro/runner"
)

func Tracer() macro.Handler {
	return macro.HandlerFunc(func(cursor macro.Context, node ast.Node) error {
		if cursor.Pre { // skip first pass
			return nil
		}

		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return nil
		}

		structDef, ok := spec.Type.(*ast.StructType)
		if !ok {
			return fmt.Errorf("only structs allowed, got type %T", spec.Type)
		}

		if structDef.Fields == nil || len(structDef.Fields.List) == 0 {
			return fmt.Errorf("given struct is empty")
		}

		f := builders.NewFileBuilder(cursor.Package.Name())
		recv := ast.NewIdent(strings.ToLower(spec.Name.Name[:1]))
		for _, field := range structDef.Fields.List {
			cb, ok := field.Type.(*ast.FuncType)
			if !ok {
				// skip fields which is not a function
				continue
			}

			// cases like
			// struct {
			//	A, B, C string
			// }
			for _, name := range field.Names {
				// skip embedded fields
				if name == nil {
					continue
				}

				f.DeclareFunction("On"+name.Name, func(method builders.FunctionBuilder) builders.FunctionBuilder {
					// Set receiver
					method = method.Recv(builders.Param(recv)(builders.RefFor(spec.Name)))

					var parameterNames []ast.Expr
					if cb.Params != nil {
						c := 0
						for _, result := range cb.Params.List {
							for i := range result.Names {
								name := ast.NewIdent(fmt.Sprintf("param%d", c))
								parameterNames = append(parameterNames, name)
								result.Names[i] = name
								c++
							}
						}
						method = method.AddParameters(cb.Params.List...)
					}

					if cb.Results != nil {
						empty := ast.NewIdent("_")
						for _, result := range cb.Results.List {
							if len(result.Names) == 0 {
								result.Names = append(result.Names, empty)
							}
						}
						method = method.AddResults(cb.Results.List...)
					}

					return method.Body(func(s builders.StatementBuilder) builders.StatementBuilder {
						inil := builders.Nil()
						fieldSelector := builders.Selector(recv, name)
						cond := builders.Or(builders.Eq(recv, inil), builders.Eq(fieldSelector, inil))

						s = s.If(nil, cond, func(ifBody builders.StatementBuilder) builders.StatementBuilder {
							return ifBody.Return()
						})

						if cb.Results != nil {
							// Return, if result tuple is not empty
							s = s.Return(builders.Call(fieldSelector, parameterNames...))
						} else {
							s = s.Expr(builders.Call(fieldSelector, parameterNames...))
						}

						return s
					})
				})
			}

		}

		cursor.AddDecls(f.Complete().Decls...)
		return nil
	})
}

func main() {
	if err := runner.Run("tracer.go", "tracer_gen.go", macro.Macros{
		"tracer": Tracer(),
	}); err != nil {
		panic(err)
	}
}
