package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

type obfuscated2 struct {
	Header  []byte
	Encrypt cipher.Stream
	Decrypt cipher.Stream
}

func createCTR(key, iv []byte) (stream cipher.Stream, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	stream = cipher.NewCTR(block, iv)
	return
}

func getDecryptInit(init [64]byte) (initRev [48]byte) {
	copy(initRev[:], init[8:56])
	// https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(initRev)-1; left < right; left, right = left+1, right-1 {
		initRev[left], initRev[right] = initRev[right], initRev[left]
	}

	return
}

func generateKeys(randSource io.Reader, secret []byte, protocol [4]byte, dc int16) (k obfuscated2, err error) {
	init, err := generateInit(randSource)
	if err != nil {
		return
	}

	// preallocate 256 bit key + 16 bit secret
	const keyLength = 32 + 16

	encryptKey := append(make([]byte, 0, keyLength), init[8:40]...)
	encryptIV := append(make([]byte, 0, 16), init[40:56]...)

	initRev := getDecryptInit(init)
	decryptKey := append(make([]byte, 0, keyLength), initRev[:32]...)
	decryptIV := append(make([]byte, 0, 16), initRev[32:48]...)
	secret = secret[0:16]

	encryptKey = crypto.SHA256(append(encryptKey, secret...))
	decryptKey = crypto.SHA256(append(decryptKey, secret...))

	k.Encrypt, err = createCTR(encryptKey, encryptIV)
	if err != nil {
		return
	}

	k.Decrypt, err = createCTR(decryptKey, decryptIV)
	if err != nil {
		return
	}

	copy(init[56:60], protocol[:])
	binary.LittleEndian.PutUint16(init[60:62], uint16(dc))

	var encryptedInit [64]byte
	k.Encrypt.XORKeyStream(encryptedInit[:], init[:])
	k.Header = make([]byte, 64)
	copy(k.Header, init[0:56])
	copy(k.Header[56:], encryptedInit[56:56+8])

	return k, nil
}

// function from https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func generateInit(randSource io.Reader) (init [64]byte, err error) {
	// init := (56 random bytes) + protocol + dc + (2 random bytes)
	for {
		_, err = io.ReadFull(randSource, init[:])
		if err != nil {
			return
		}

		if init[0] == 0xef {
			continue
		}

		firstInt := binary.LittleEndian.Uint32(init[0:4])
		if firstInt == 0x44414548 ||
			firstInt == 0x54534f50 ||
			firstInt == 0x20544547 ||
			firstInt == 0x4954504f ||
			firstInt == 0x02010316 ||
			firstInt == 0xdddddddd ||
			firstInt == 0xeeeeeeee {
			continue
		}

		if secondInt := binary.LittleEndian.Uint32(init[4:8]); secondInt == 0 {
			continue
		}

		break
	}

	return init, nil
}

func (o obfuscated2) Writer(w io.Writer) io.Writer {
	return obfuscated2IO{obfuscated2: o, w: w}
}

func (o obfuscated2) Reader(r io.Reader) io.Reader {
	return obfuscated2IO{obfuscated2: o, r: r}
}

type obfuscated2IO struct {
	obfuscated2
	w io.Writer
	r io.Reader
}

func (o obfuscated2IO) Write(b []byte) (n int, err error) {
	o.Encrypt.XORKeyStream(b, b)
	return o.w.Write(b)
}

func (o obfuscated2IO) Read(b []byte) (n int, err error) {
	n, err = o.r.Read(b)
	if err != nil {
		return
	}
	o.Decrypt.XORKeyStream(b, b)
	return
}

// MTProxyObfuscated2 implements MTProxy transport wrapper.
type MTProxyObfuscated2 struct {
	Codec  TaggedCodec
	DC     int16
	Secret []byte
	obs    obfuscated2
}

// WriteHeader sends protocol tag.
func (o *MTProxyObfuscated2) WriteHeader(w io.Writer) (err error) {
	o.obs, err = generateKeys(rand.Reader, o.Secret, o.Codec.ObfuscatedTag(), o.DC)
	if err != nil {
		return err
	}

	if _, err := w.Write(o.obs.Header); err != nil {
		return xerrors.Errorf("write obfuscated header: %w", err)
	}

	return
}

// ReadHeader reads protocol tag.
func (o *MTProxyObfuscated2) ReadHeader(r io.Reader) error {
	return errors.New("server side not implemented yet")
}

// Write encode to writer message from given buffer.
func (o *MTProxyObfuscated2) Write(w io.Writer, b *bin.Buffer) error {
	return o.Codec.Write(o.obs.Writer(w), b)
}

// Read fills buffer with received message.
func (o *MTProxyObfuscated2) Read(r io.Reader, b *bin.Buffer) error {
	return o.Codec.Read(o.obs.Reader(r), b)
}
