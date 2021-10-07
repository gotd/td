package gen

// This file defines how to generate templates and example
// generated files.

//go:generate go run github.com/nnqq/td/cmd/gotdgen --doc "https://localhost:80/doc" --clean --package td --target example --schema _testdata/example.tl --server
