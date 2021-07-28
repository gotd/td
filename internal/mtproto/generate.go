package mtproto

//go:generate go run github.com/gotd/td/cmd/rsagen -f _data/public_keys.pem -pkg mtproto -var vendoredKeys -o vendored_keys.go
//go:generate go run github.com/gotd/td/cmd/rsagen -f _data/public_keys.pem -pkg mtproto -var vendoredKeys -test TestVendoredKeys -o vendored_keys_test.go -exec test.tmpl
