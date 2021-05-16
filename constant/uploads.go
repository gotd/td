package constant

// https://core.telegram.org/api/files#uploading-files
const (
	// UploadMaxSmallSize is maximum size of small file.
	//
	// Use upload.saveBigFilePart in case the full size of the file is more than 10 MB
	// and upload.saveFilePart for smaller files
	UploadMaxSmallSize = 10 * 1024 * 1024 // 10 MB
	// UploadMaxParts is maximum parts count.
	//
	// Each part should have a sequence number, file_part, with a value ranging from 0 to 3,999.
	UploadMaxParts = 3999
	// UploadPadding is part size padding.
	//
	// `part_size % 1024 = 0` (divisible by 1KB)
	UploadPadding = 1024
	// UploadMaxPartSize is maximum size of single part.
	//
	// `524288 % part_size = 0` (512KB must be evenly divisible by part_size)
	UploadMaxPartSize = 524288
)
