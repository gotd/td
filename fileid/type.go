package fileid

//go:generate go run -modfile=../_tools/go.mod golang.org/x/tools/cmd/stringer -type=Type

// Type represents file_id type.
type Type int

const (
	// Thumbnail is Thumbnail file type.
	Thumbnail Type = iota
	// ProfilePhoto is ProfilePhoto file type.
	ProfilePhoto
	// Photo is Photo file type.
	Photo
	// Voice is Voice file type.
	Voice
	// Video is Video file type.
	Video
	// Document is Document file type.
	Document
	// Encrypted is Encrypted file type.
	Encrypted
	// Temp is Temp file type.
	Temp
	// Sticker is Sticker file type.
	Sticker
	// Audio is Audio file type.
	Audio
	// Animation is Animation file type.
	Animation
	// EncryptedThumbnail is EncryptedThumbnail file type.
	EncryptedThumbnail
	// Wallpaper is Wallpaper file type.
	Wallpaper
	// VideoNote is VideoNote file type.
	VideoNote
	// SecureRaw is SecureRaw file type.
	SecureRaw
	// Secure is Secure file type.
	Secure
	// Background is Background file type.
	Background
	lastType
)
