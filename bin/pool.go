package bin

import "sync"

// Pool is a bin.Buffer pool.
type Pool struct {
	pool sync.Pool
}

// NewPool creates new Pool.
// Length is initial buffer length.
func NewPool(length int) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() interface{} {
				var r []byte
				if length > 0 {
					r = make([]byte, 0, length)
				}
				return &Buffer{Buf: r}
			},
		},
	}
}

// Put returns buffer to pool.
func (b *Pool) Put(buf *Buffer) {
	b.pool.Put(buf)
}

// Get takes buffer from pool.
func (b *Pool) Get() *Buffer {
	buf := b.pool.Get().(*Buffer)
	buf.Reset()
	return buf
}

// GetSize takes buffer with given size from pool.
func (b *Pool) GetSize(length int) *Buffer {
	buf := b.Get()
	buf.ResetN(length)

	return buf
}
