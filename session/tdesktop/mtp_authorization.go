package tdesktop

import (
	"github.com/ogen-go/errors"

	"github.com/gotd/td/internal/crypto"
)

// MTPAuthorization is a Telegram Desktop storage structure which stores MTProto session info.
type MTPAuthorization struct {
	// UserID is a Telegram user ID.
	UserID int64
	// MainDC is a main DC ID of this user.
	MainDC int
	// Key is a map of keys per DC ID.
	Keys map[int]crypto.Key // DC ID -> Key
}

func readKey(r *reader, k *crypto.Key) (uint32, error) {
	dcID, err := r.readUint32()
	if err != nil {
		return 0, errors.Wrap(err, "read DC ID")
	}

	if err := r.consumeN(k[:], 256); err != nil {
		return 0, errors.Wrap(err, "read auth key")
	}

	return dcID, nil
}

func (m *MTPAuthorization) deserialize(r *reader) error {
	id, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read dbi ID")
	}
	if id != dbiMtpAuthorization {
		return errors.Errorf("unexpected id %d", id)
	}

	if err := r.skip(4); err != nil {
		return errors.Wrap(err, "read mainLength")
	}

	legacyUserID, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read legacyUserID")
	}
	legacyMainDCID, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read legacyMainDCID")
	}
	if (uint64(legacyUserID)<<32)|uint64(legacyMainDCID) == kWideIdsTag {
		userID, err := r.readUint64()
		if err != nil {
			return errors.Wrap(err, "read userID")
		}
		mainDC, err := r.readUint32()
		if err != nil {
			return errors.Wrap(err, "read mainDcID")
		}

		m.UserID = int64(userID)
		m.MainDC = int(mainDC)
	} else {
		m.UserID = int64(legacyUserID)
		m.MainDC = int(legacyMainDCID)
	}

	keys, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read keys length")
	}

	if m.Keys == nil {
		m.Keys = make(map[int]crypto.Key, keys)
	}
	for i := 0; i < int(keys); i++ {
		var key crypto.Key
		dcID, err := readKey(r, &key)
		if err != nil {
			return errors.Wrapf(err, "read key %d", i)
		}
		// FIXME(tdakkota): what if there is more than one session per DC?
		m.Keys[int(dcID)] = key
	}

	return nil
}
