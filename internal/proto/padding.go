package proto

const padding = 1024

func paddedLen(l int) int {
	n := padding * (l / padding)
	if n < l {
		n += padding
	}
	return n
}

func paddingRequired(l int) int {
	return paddedLen(l) - l
}
