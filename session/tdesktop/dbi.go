package tdesktop

import (
	"io"
	"math/bits"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

//nolint:deadcode,unused,varcheck
const (
	dbiKey               = 0x00
	dbiUser              = 0x01
	dbiDcOptionOldOld    = 0x02
	dbiChatSizeMax       = 0x03
	dbiMutePeer          = 0x04
	dbiSendKey           = 0x05
	dbiAutoStart         = 0x06
	dbiStartMinimized    = 0x07
	dbiSoundNotify       = 0x08
	dbiWorkMode          = 0x09
	dbiSeenTrayTooltip   = 0x0a
	dbiDesktopNotify     = 0x0b
	dbiAutoUpdate        = 0x0c
	dbiLastUpdateCheck   = 0x0d
	dbiWindowPosition    = 0x0e
	dbiConnectionTypeOld = 0x0f
	// 0x10 reserved
	dbiDefaultAttach     = 0x11
	dbiCatsAndDogs       = 0x12
	dbiReplaceEmojis     = 0x13
	dbiAskDownloadPath   = 0x14
	dbiDownloadPathOld   = 0x15
	dbiScale             = 0x16
	dbiEmojiTabOld       = 0x17
	dbiRecentEmojiOldOld = 0x18
	dbiLoggedPhoneNumber = 0x19
	dbiMutedPeers        = 0x1a
	// 0x1b reserved
	dbiNotifyView              = 0x1c
	dbiSendToMenu              = 0x1d
	dbiCompressPastedImage     = 0x1e
	dbiLangOld                 = 0x1f
	dbiLangFileOld             = 0x20
	dbiTileBackground          = 0x21
	dbiAutoLock                = 0x22
	dbiDialogLastPath          = 0x23
	dbiRecentEmojiOld          = 0x24
	dbiEmojiVariantsOld        = 0x25
	dbiRecentStickers          = 0x26
	dbiDcOptionOld             = 0x27
	dbiTryIPv6                 = 0x28
	dbiSongVolume              = 0x29
	dbiWindowsNotificationsOld = 0x30
	dbiIncludeMuted            = 0x31
	dbiMegagroupSizeMax        = 0x32
	dbiDownloadPath            = 0x33
	dbiAutoDownload            = 0x34
	dbiSavedGifsLimit          = 0x35
	dbiShowingSavedGifsOld     = 0x36
	dbiAutoPlay                = 0x37
	dbiAdaptiveForWide         = 0x38
	dbiHiddenPinnedMessages    = 0x39
	dbiRecentEmoji             = 0x3a
	dbiEmojiVariants           = 0x3b
	dbiDialogsMode             = 0x40
	dbiModerateMode            = 0x41
	dbiVideoVolume             = 0x42
	dbiStickersRecentLimit     = 0x43
	dbiNativeNotifications     = 0x44
	dbiNotificationsCount      = 0x45
	dbiNotificationsCorner     = 0x46
	dbiThemeKey                = 0x47
	dbiDialogsWidthRatioOld    = 0x48
	dbiUseExternalVideoPlayer  = 0x49
	dbiDcOptions               = 0x4a
	dbiMtpAuthorization        = 0x4b
	dbiLastSeenWarningSeenOld  = 0x4c
	dbiAuthSessionSettings     = 0x4d
	dbiLangPackKey             = 0x4e
	dbiConnectionType          = 0x4f
	dbiStickersFavedLimit      = 0x50
	dbiSuggestStickersByEmoji  = 0x51

	dbiEncryptedWithSalt = 333
	dbiEncrypted         = 444
)

type qtReader struct {
	buf bin.Buffer
}

func (r *qtReader) subArray() (qtReader, error) {
	length, err := r.readInt32()
	if err != nil {
		return qtReader{}, errors.Wrap(err, "read length")
	}
	sub := bin.Buffer{Buf: r.buf.Buf}
	if err := r.skip(int(length)); err != nil {
		return qtReader{}, io.ErrUnexpectedEOF
	}

	sub.Buf = sub.Buf[:length]
	return qtReader{buf: sub}, err
}

func (r *qtReader) readUint64() (uint64, error) {
	u, err := r.buf.Uint64()
	return bits.ReverseBytes64(u), err
}

func (r *qtReader) readUint32() (uint32, error) {
	u, err := r.buf.Uint32()
	return bits.ReverseBytes32(u), err
}

func (r *qtReader) readInt32() (int32, error) {
	v, err := r.readUint32()
	return int32(v), err
}

func (r *qtReader) readString() (string, error) {
	sz, err := r.readInt32()
	if err != nil {
		return "", err
	}
	size := int(sz)
	switch {
	case size < 0:
		return "", &bin.InvalidLengthError{
			Length: size,
			Where:  "QString",
		}
	case size >= r.buf.Len():
		return "", io.ErrUnexpectedEOF
	}
	s := string(r.buf.Buf[:size])
	r.buf.Skip(size)
	return s, nil
}

func (r *qtReader) readBytes() ([]byte, error) {
	sz, err := r.readInt32()
	if err != nil {
		return nil, err
	}
	size := int(sz)
	switch {
	case size < 0:
		return nil, &bin.InvalidLengthError{
			Length: size,
			Where:  "QString",
		}
	case size > r.buf.Len():
		return nil, io.ErrUnexpectedEOF
	}
	s := append([]byte(nil), r.buf.Buf[:size]...)
	r.buf.Skip(size)
	return s, nil
}

func (r *qtReader) consumeN(target []byte, n int) error {
	return r.buf.ConsumeN(target, n)
}

func (r *qtReader) skip(n int) error {
	if r.buf.Len() < n {
		return io.ErrUnexpectedEOF
	}
	r.buf.Skip(n)
	return nil
}
