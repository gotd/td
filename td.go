// Package td implements MTProto encoding and decoding.
package td

//go:generate go run github.com/gotd/td/cmd/gotdgen --doc "https://core.telegram.org/" --clean --package tg --target tg --schema _schema/telegram.tl
//go:generate go run github.com/gotd/td/cmd/gotdgen --doc "https://core.telegram.org/" --clean --package e2e --target tg/e2e --schema _schema/encrypted.tl
