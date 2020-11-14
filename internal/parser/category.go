package parser

// Category of Definition.
type Category byte

//go:generate go run golang.org/x/tools/cmd/stringer -type=Category

const (
	CategoryType Category = iota
	CategoryFunction
)
