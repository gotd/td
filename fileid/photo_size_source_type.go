package fileid

//go:generate go run -modfile=../_tools/go.mod golang.org/x/tools/cmd/stringer -type=PhotoSizeSourceType

// PhotoSizeSourceType represents photo_size_source type.
type PhotoSizeSourceType int

const (
	// PhotoSizeSourceLegacy is Legacy type.
	PhotoSizeSourceLegacy PhotoSizeSourceType = iota
	// PhotoSizeSourceThumbnail is Thumbnail type.
	PhotoSizeSourceThumbnail
	// PhotoSizeSourceDialogPhotoSmall is DialogPhotoSmall type.
	PhotoSizeSourceDialogPhotoSmall
	// PhotoSizeSourceDialogPhotoBig is DialogPhotoBig type.
	PhotoSizeSourceDialogPhotoBig
	// PhotoSizeSourceStickerSetThumbnail is StickerSetThumbnail type.
	PhotoSizeSourceStickerSetThumbnail
	// PhotoSizeSourceFullLegacy is FullLegacy type.
	PhotoSizeSourceFullLegacy
	// PhotoSizeSourceDialogPhotoSmallLegacy is DialogPhotoSmallLegacy type.
	PhotoSizeSourceDialogPhotoSmallLegacy
	// PhotoSizeSourceDialogPhotoBigLegacy is DialogPhotoBigLegacy type.
	PhotoSizeSourceDialogPhotoBigLegacy
	// PhotoSizeSourceStickerSetThumbnailLegacy is StickerSetThumbnailLegacy type.
	PhotoSizeSourceStickerSetThumbnailLegacy
	// PhotoSizeSourceStickerSetThumbnailVersion is StickerSetThumbnailVersion type.
	PhotoSizeSourceStickerSetThumbnailVersion
	lastPhotoSizeSourceType
)
