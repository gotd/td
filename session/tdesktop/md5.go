package tdesktop

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func tdesktopMD5(s string) string {
	hash := md5.Sum([]byte(s))
	for i := range hash {
		hash[i] = hash[i] << 4 | hash[i] >> 4
	}
	hexed := hex.EncodeToString(hash[:])
	return strings.ToUpper(hexed)
}

func fileKey(s string) string {
	return tdesktopMD5(s)[:16]
}
