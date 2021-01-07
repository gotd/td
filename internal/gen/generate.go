package gen

// This file defines how to generate templates and example
// generated files.

// Templates should be first.
//go:generate go run github.com/gotd/td/internal/tools/templategen -pkg=internal -o=internal/bindata.go ./_template/...

//go:generate go run github.com/gotd/td/cmd/gotdgen --doc "https://localhost:80/doc" --clean --package td --target example --schema _testdata/example.tl
