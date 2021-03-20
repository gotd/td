package updates

import (
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

type channelKey tg.InputChannel

func (c channelKey) String() string {
	var b [32]byte
	s := strings.Builder{}
	s.Grow(48)

	s.Write(strconv.AppendInt(b[:0], int64(c.ChannelID), 10))
	s.WriteByte('_')
	s.Write(strconv.AppendInt(b[:0], c.AccessHash, 10))

	return s.String()
}

func (c *channelKey) Parse(s string) error {
	idx := strings.Index(s, "_")
	switch {
	case idx < 0:
		return xerrors.Errorf("bad %q key, expected '_'", s)
	case idx == len(s)-1:
		return xerrors.Errorf("bad %q key, no access hash", s)
	}

	id, err := strconv.Atoi(s[:idx])
	if err != nil {
		return xerrors.Errorf("parse id: %w", err)
	}

	hash, err := strconv.Atoi(s[idx+1:])
	if err != nil {
		return xerrors.Errorf("parse access hash: %w", err)
	}

	c.ChannelID = id
	c.AccessHash = int64(hash)
	return nil
}
