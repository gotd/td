package faketls

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	a := require.New(t)
	r := record{
		Type:    RecordTypeApplication,
		Version: Version10Bytes,
		Data:    []byte(`abcd`),
	}
	buf := bytes.NewBuffer(nil)

	_, err := writeRecord(buf, r)
	a.NoError(err)
	r2, err := readRecord(buf)
	a.NoError(err)
	a.Equal(r, r2)
}
