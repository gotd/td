package syncio

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufWriterAt(t *testing.T) {
	a := require.New(t)
	f := BufWriterAt{}
	var err error

	testData := []byte{1, 2, 3, 4, 5}
	buf := make([]byte, len(testData))
	offset := len(testData)
	zeros := bytes.Repeat([]byte{0}, offset)

	_, err = f.WriteAt([]byte{1}, -1)
	a.Error(err)

	_, err = f.ReadAt([]byte{}, -1)
	a.Error(err)

	_, err = f.ReadAt([]byte{1}, 0)
	a.NoError(err)

	// 5 zeros + testdata
	_, err = f.WriteAt(testData, int64(offset))
	a.NoError(err)
	a.Equal(testData, f.buf[offset:])
	a.Equal(zeros, f.buf[:offset])
	_, err = f.ReadAt(buf, int64(offset))
	a.NoError(err)
	a.Equal(testData, buf)

	// 2 * testdata
	_, err = f.WriteAt(testData, 0)
	a.NoError(err)
	a.Equal(bytes.Repeat(testData, 2), f.buf)
	_, err = f.ReadAt(buf, 0)
	a.NoError(err)
	a.Equal(testData, buf)

	// 3 * testdata
	_, err = f.WriteAt(testData, int64(offset*2))
	a.NoError(err)
	a.Equal(bytes.Repeat(testData, 3), f.buf)
	_, err = f.ReadAt(buf, int64(offset*2))
	a.NoError(err)
	a.Equal(testData, buf)

	// 3 * testdata + 5 zeros + testdata
	_, err = f.WriteAt(testData, int64(offset*4))
	a.NoError(err)
	a.Equal(bytes.Repeat(testData, 3), f.buf[:3*offset])
	a.Equal(zeros, f.buf[3*offset:4*offset])
	a.Equal(testData, f.buf[4*offset:])
	_, err = f.ReadAt(buf, int64(offset*4))
	a.NoError(err)
	a.Equal(testData, buf)

	// 5 * testdata
	_, err = f.WriteAt(testData, int64(offset*3))
	a.NoError(err)
	a.Equal(bytes.Repeat(testData, 5), f.buf)
}
