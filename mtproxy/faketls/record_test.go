package faketls

import (
	"bytes"
	"strings"
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

func TestRecord_AllowsLargeApplicationPayload(t *testing.T) {
	a := require.New(t)

	payload := []byte(strings.Repeat("a", 20_000))
	r := record{
		Type:    RecordTypeApplication,
		Version: Version12Bytes,
		Data:    payload,
	}
	buf := bytes.NewBuffer(nil)

	_, err := writeRecord(buf, r)
	a.NoError(err)

	r2, err := readRecord(buf)
	a.NoError(err)
	a.Equal(r, r2)
}
