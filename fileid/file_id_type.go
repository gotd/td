package fileid

//go:generate go run -modfile=../_tools/go.mod golang.org/x/tools/cmd/stringer -type=Type

// Type represents file_id type.
type Type int

const (
	Thumbnail Type = iota
	ProfilePhoto
	Photo
	Voice
	Video
	Document
	Encrypted
	Temp
	Sticker
	Audio
	Animation
	EncryptedThumbnail
	Wallpaper
	VideoNote
	SecureRaw
	Secure
	Background
	lastType
)
