package message

var (
	_ MediaOption = (*PhotoExternalBuilder)(nil)
	_ MediaOption = (*DocumentExternalBuilder)(nil)
)

var (
	_ MultiMediaOption = (*UploadedPhotoBuilder)(nil)
	_ MultiMediaOption = (*UploadedDocumentBuilder)(nil)
	_ MultiMediaOption = (*VideoDocumentBuilder)(nil)
	_ MultiMediaOption = (*AudioDocumentBuilder)(nil)
	_ MultiMediaOption = (*SearchDocumentBuilder)(nil)
)
