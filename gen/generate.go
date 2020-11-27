package gen

// Templates should be first.
//go:generate go run github.com/go-bindata/go-bindata/go-bindata -pkg=internal -o=internal/bindata.go -mode=420 -modtime=1 ./_template/...

//go:generate go run github.com/ernado/td/gen/cmd/gotdgen --clean --package td --target example --schema _testdata/Error.tl
