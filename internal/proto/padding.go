package proto

const padding = 16

func paddedLen(l int) int {
	n := padding * (l / padding)
	if n < l {
		n += padding
	}
	return n
}

func paddingRequired(l int) int {
	return 16 + (16 - (l % 16))
}
