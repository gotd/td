package tdesktop

import (
	"github.com/go-faster/errors"

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

// IPv6 denotes that the specified IP is an IPv6 address.
func (m MTPDCOption) IPv6() bool {
	return m.Flags.Has(0)
}

// MediaOnly denotes that this DC should only be used to download or upload files.
func (m MTPDCOption) MediaOnly() bool {
	return m.Flags.Has(1)
}

// TCPOOnly denotes that this DC only supports connection with transport obfuscation.
func (m MTPDCOption) TCPOOnly() bool {
	return m.Flags.Has(2)
}

// CDN denotes that this is a CDN DC.
func (m MTPDCOption) CDN() bool {
	return m.Flags.Has(3)
}

// Static denotes that this IP should be used when connecting through a proxy.
func (m MTPDCOption) Static() bool {
	return m.Flags.Has(4)
}

func (m *MTPDCOption) deserialize(r *qtReader, version int32) error {
	id, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read id")
	}
	m.ID = id

	flags, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read flags")
	}
	m.Flags = bin.Fields(flags)

	port, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read port")
	}
	m.Port = port

	const maxIPSize = 45
	ip, err := r.readString()
	if err != nil {
		return errors.Wrap(err, "read ip")
	}
	if l := len(ip); l > maxIPSize {
		return errors.Errorf("too big IP string (%d > %d)", l, maxIPSize)
	}
	m.IP = ip

	if version > 0 {
		const maxSecretSize = 32
		secret, err := r.readBytes()
		if err != nil {
			return errors.Wrap(err, "read secret")
		}
		if l := len(secret); l > maxSecretSize {
			return errors.Errorf("too big DC secret (%d > %d)", l, maxSecretSize)
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
		return errors.Wrap(err, "read version")
	}

	var version int32
	if minusVersion < 0 {
		version = -minusVersion
	}

	var count int32
	if version > 0 {
		c, err := r.readInt32()
		if err != nil {
			return errors.Wrap(err, "read count")
		}
		count = c
	} else {
		count = minusVersion
	}

	for i := 0; i < int(count); i++ {
		var o MTPDCOption
		if err := o.deserialize(r, version); err != nil {
			return errors.Errorf("read option %d: %w", i, err)
		}
		m.Options = append(m.Options, o)
	}

	// TODO(tdakkota): Read CDN keys.
	return nil
}
