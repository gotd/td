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

// CachedQuery is a RPC cacheable query helper.
type CachedQuery struct {
	// Name of struct to generate.
	Name string
	// MethodName is name of method of tg.Client.
	MethodName string
	// RequestName is name of request struct.
	RequestName string
	// ManualHash determines whether hash must be computed using
	// hand-written function computeHash or not.
	// Need to resolve case when Telegram does not return hash with result.
	ManualHash bool
	// RequestParams contains additional params to send.
	RequestParams []Param
	// ResultName is name of result type.
	ResultName string
	// NotModifiedName is name of NotModified result type.
	NotModifiedName string
}

// Config is codegeneration config to use.
type Config struct {
	// Query helpers to generate.
	Queries []CachedQuery
	// ResultName package name
	Package string
}
