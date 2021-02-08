// Package downloader contains downloading files helpers.
package downloader

import (
	"io"

	"github.com/gotd/td/tg"
)

// Download represents Telegram file download.
type Download struct {
	file tg.InputFileLocationClass
	cdn  bool

	output io.Writer
}

// NewDownload creates new Download struct.
func NewDownload(file tg.InputFileLocationClass, cdn bool, output io.Writer) Download {
	return Download{file: file, cdn: cdn, output: output}
}
