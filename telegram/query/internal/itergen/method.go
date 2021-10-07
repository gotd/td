package main

import (
	"go/types"
	"sort"
	"strings"

	"github.com/nnqq/td/telegram/query/internal/genutil"
)

// Param represents request parameter.
type Param struct {
	// Name to use in function declaration.
	Name string
	// OriginalName in struct definition.
	OriginalName string
	// Go type.
	Type string
}

func varToParam(field *types.Var) Param {
	fieldName := field.Name()
	fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
	return Param{
		Name:         fieldName,
		OriginalName: field.Name(),
		Type:         genutil.PrintType(field.Type()),
	}
}

func sortParams(p []Param) []Param {
	sort.SliceStable(p, func(i, j int) bool {
		return p[i].Name < p[j].Name
	})

	return p
}

// SpecialCaseChain represents special request parameter setter.
type SpecialCaseChain struct {
	// ConstructorName to use in function body.
	ConstructorName string
	// ConstructorType to use in function body.
	ConstructorType string
	// Field of request struct.
	Field Param
	// Args is a slice of arguments. May be empty.
	Args []Param
}

// RequestArgument represents request parameter passed by iterator.
type RequestArgument struct {
	// Arg describes argument.
	Arg Param
	// Chain is flag to generate builder chain setter.
	Chain bool
	// RequiredByIter is flag to generate pass to iterator constructor.
	RequiredByIter bool
}

// Method is a RPC method.
type Method struct {
	// Name to use in function declaration.
	Name string
	// OriginalName is name of method of tg.Client.
	OriginalName string
	// RequestName is name of request struct.
	RequestName string
	// ResultName is name of result type.
	ResultName string
	// RequiredParams is a required params for query builder.
	RequiredParams []Param
	// AdditionalMapping is names of field from iterator.
	// Some type doesn't have AddOffset for example, so we customize mapping here.
	AdditionalMapping []RequestArgument

	// SpecialCase is a slice of special case chains.
	// Like tg.MessagesFilterClass constructor field setters.
	SpecialCase []SpecialCaseChain

	// Other parameters of request to pass it in constructor.
	AdditionalParams []Param

	// IteratorName is name of iterator to build.
	IteratorName string
	// ElemName  is name of iterator elem.
	ElemName string
}

// Config is codegeneration config to use.
type Config struct {
	// Methods to generate helpers and query builders.
	Methods []Method
	// ResultName package name
	Package string
	// ResultName is name of result type.
	ResultName string
	// RequestFields is a slice of request struct fields.
	RequestFields []Param
}
