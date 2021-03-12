package query

// Hasher implements Telegram pagination hash counting.
//
// See https://core.telegram.org/api/offsets#hash-generation.
type Hasher struct {
	state uint32
}

// Reset resets the Hasher to its initial state.
func (h *Hasher) Reset() {
	h.state = 0
}

// Update performs state change using given value.
func (h *Hasher) Update(value uint32) {
	h.state = (h.state * 20261) + value
}

// Update64 performs state change using given 64-bit value.
func (h *Hasher) Update64(value uint64) {
	h.Update(uint32(value >> 32))
	h.Update(uint32(value & 0xFFFFFFFF))
}

// Sum returns final sum.
func (h *Hasher) Sum() int32 {
	r := int32(h.state & 0x7FFFFFFF)
	return r
}
