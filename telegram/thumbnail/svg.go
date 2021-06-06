package thumbnail

import (
	"errors"
	"strconv"
)

// DecodePath decodes vector thumbnail from thumbnail bytes (e.g. tg.PhotoPathSize with type "j").
//
// See DecodePathTo.
func DecodePath(data []byte) ([]byte, error) {
	return DecodePathTo(data, nil)
}

// ErrInvalidPath denotes that vector thumbnail is invalid and cannot be decoded.
var ErrInvalidPath = errors.New("invalid path")

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
func DecodePathTo(data, to []byte) ([]byte, error) {
	const (
		lookup = "AACAAAAHAAALMAAAQASTAVAAAZaacaaaahaaalmaaaqastava.az0123456789-,"
	)
	to = append(to, 'M')
	for _, num := range data {
		if num >= 128+64 {
			r, ok := safeLookup(lookup, int(num-128-64))
			if !ok {
				return nil, ErrInvalidPath
			}
			to = append(to, r)
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
	return to, nil
}

func safeLookup(s string, idx int) (byte, bool) {
	if idx < 0 || len(s) <= idx {
		return 0, false
	}
	return s[idx], true
}
