package thumbnail

import (
	"strconv"
)

// DecodePath decodes vector thumbnail from thumbnail bytes (e.g. tg.PhotoPathSize with type "j").
//
// See DecodePathTo.
func DecodePath(data []byte) []byte {
	return DecodePathTo(data, nil)
}

// DecodePathTo decodes vector thumbnail to given slice.
//
// Returned path will contain the actual SVG path that can
// be directly inserted in the d attribute of an svg <path> element:
//
//	<?xml version="1.0" encoding="utf-8"?>
//	<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
//	   viewBox="0 0 512 512" xml:space="preserve">
//	<path d="{$path}"/>
//	</svg>
//
// See https://core.telegram.org/api/files#vector-thumbnails.
func DecodePathTo(data, to []byte) []byte {
	const (
		lookup = "AACAAAAHAAALMAAAQASTAVAAAZaacaaaahaaalmaaaqastava.az0123456789-,"
	)
	to = append(to, 'M')
	for _, num := range data {
		if num >= 128+64 {
			to = append(to, lookup[int(num-128-64)])
		} else {
			if num >= 128 {
				to = append(to, ',')
			} else if num >= 64 {
				to = append(to, '-')
			}
			to = strconv.AppendInt(to, int64(num&63), 10)
		}
	}
	to = append(to, 'z')
	return to
}
