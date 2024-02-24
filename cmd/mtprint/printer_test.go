package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tg"
)

func Test_readAndPrint(t *testing.T) {
	c := codec.Intermediate{}

	input := &bytes.Buffer{}
	buf := &bin.Buffer{}

	objects := []bin.Object{
		&mt.RPCResult{},
		&mt.RPCError{},
		&tg.CodeSettings{},
	}
	for _, o := range objects {
		buf.Reset()
		require.NoError(t, o.Encode(buf))
		require.NoError(t, c.Write(input, buf))
	}

	output := &bytes.Buffer{}
	require.NoError(t, NewPrinter(input, formats("go"), c).Print(output))
	out := output.String()
	require.Contains(t, out, "RPCResult")
	require.Contains(t, out, "RPCError")
	require.Contains(t, out, "CodeSettings")
}
