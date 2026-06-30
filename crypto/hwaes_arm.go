//go:build arm && cgo && (linux || android)

package crypto

/*
#cgo CFLAGS: -O3 -march=armv8-a+crypto -mfpu=crypto-neon-fp-armv8
#include <stdint.h>
int aes256_ige_decrypt(const uint8_t *key, const uint8_t *iv,
                       uint8_t *dst, const uint8_t *src, int len);
*/
import "C"
import "unsafe"

// hwIGEDecrypt decrypts AES-256-IGE using the ARMv8 Cryptography Extension when
// the running CPU exposes it. It returns false when the hardware path is not
// available or the input is invalid, allowing the caller to use the Go fallback.
func hwIGEDecrypt(key, iv, dst, src []byte) bool {
	if len(src) == 0 || len(src)%16 != 0 {
		return false
	}
	if len(key) < 32 || len(iv) < 32 || len(dst) < len(src) {
		return false
	}
	rc := C.aes256_ige_decrypt(
		(*C.uint8_t)(unsafe.Pointer(&key[0])),
		(*C.uint8_t)(unsafe.Pointer(&iv[0])),
		(*C.uint8_t)(unsafe.Pointer(&dst[0])),
		(*C.uint8_t)(unsafe.Pointer(&src[0])),
		C.int(len(src)),
	)
	return rc == 0
}
