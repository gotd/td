//go:build !arm || !cgo || (!linux && !android)

package crypto

func hwIGEDecrypt(key, iv, dst, src []byte) bool { return false }
