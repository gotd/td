package ascii

// IsDigit reports whether the rune is a decimal digit.
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
