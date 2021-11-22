package fileid

//go:generate go run -modfile=../_tools/go.mod golang.org/x/tools/cmd/stringer -type=PhotoSizeSourceType

// PhotoSizeSourceType represents photo_size_source type.
type PhotoSizeSourceType int

const (
	PhotoSizeSourceLegacy PhotoSizeSourceType = iota
	PhotoSizeSourceThumbnail
	PhotoSizeSourceDialogPhotoSmall
	PhotoSizeSourceDialogPhotoBig
	PhotoSizeSourceStickerSetThumbnail
	PhotoSizeSourceFullLegacy
	PhotoSizeSourceDialogPhotoSmallLegacy
	PhotoSizeSourceDialogPhotoBigLegacy
	PhotoSizeSourceStickerSetThumbnailLegacy
	PhotoSizeSourceStickerSetThumbnailVersion
	lastPhotoSizeSourceType
)
