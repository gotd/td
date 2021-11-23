package fileid

import "bytes"

func rleEncode(s []byte) (r []byte) {
	var count byte
	for _, cur := range s {
		if cur == 0 {
			count++
			continue
		}

		if count > 0 {
			r = append(r, 0, count)
			count = 0
		}
		r = append(r, cur)
	}
	if count > 0 {
		r = append(r, 0, count)
	}

	return r
}

func rleDecode(s []byte) (r []byte) {
	var last []byte
	for _, cur := range s {
		if string(last) == string(rune(0)) {
			r = append(r, bytes.Repeat(last, int(cur))...)
			last = nil
		} else {
			r = append(r, last...)
			last = []byte{cur}
		}
	}
	r = append(r, last...)
	return r
}
