package uploader

import "sync"

type bufferPool struct {
	pool sync.Pool
}

func newBufferPool(partSize int) *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				r := make([]byte, partSize)
				return &r
			},
		},
	}
}

func (b *bufferPool) Put(buf []byte) {
	b.pool.Put(&buf)
}

func (b *bufferPool) Get() []byte {
	return *b.pool.Get().(*[]byte)
}

func (b *bufferPool) GetSize(n int) []byte {
	get := b.Get()
	return append(get[:0], make([]byte, n)...)
}
