package tdesktop

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// MTPDCOption is a Telegram Desktop storage structure which stores DC info.
//
// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/mtproto/mtproto_dc_options.h.
type MTPDCOption struct {
	ID     int32
	Flags  bin.Fields
	Port   int32
	IP     string
	Secret []byte
}

func (m *MTPDCOption) deserialize(r *qtReader, version int32) error {
	id, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read id: %w", err)
	}
	m.ID = id

	fields, err := r.readUint32()
	if err != nil {
		return xerrors.Errorf("read flags: %w", err)
	}
	m.Flags = bin.Fields(fields)

	port, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read port: %w", err)
	}
	m.Port = port

	const kMaxIpSize = 45
	ip, err := r.readString()
	if err != nil {
		return xerrors.Errorf("read ip: %w", err)
	}
	if l := len(ip); l > kMaxIpSize {
		return xerrors.Errorf("too big IP string (%d > %d)", l, kMaxIpSize)
	}
	m.IP = ip

	if version > 0 {
		const kMaxSecretSize = 32
		secret, err := r.readBytes()
		if err != nil {
			return xerrors.Errorf("read secret: %w", err)
		}
		if l := len(secret); l > kMaxSecretSize {
			return xerrors.Errorf("too big DC secret (%d > %d)", l, kMaxSecretSize)
		}
		m.Secret = secret
	}

	return nil
}

// MTPDCOptions is a Telegram Desktop storage structure which stores DCs info.
//
// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/mtproto/mtproto_dc_options.cpp#L479.
type MTPDCOptions struct {
	Options []MTPDCOption
}

func (m *MTPDCOptions) deserialize(r *qtReader) error {
	minusVersion, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read version: %w", err)
	}

	var version int32
	if minusVersion < 0 {
		version = -minusVersion
	}

	var count int32
	if version > 0 {
		c, err := r.readInt32()
		if err != nil {
			return xerrors.Errorf("read count: %w", err)
		}
		count = c
	} else {
		count = minusVersion
	}

	for i := 0; i < int(count); i++ {
		var o MTPDCOption
		if err := o.deserialize(r, version); err != nil {
			return xerrors.Errorf("read option %d: %w", i, err)
		}
		m.Options = append(m.Options, o)
	}

	// TODO(tdakkota): Read CDN keys.
	return nil
}
