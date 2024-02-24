package tdesktop

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
)

// MTPConfigEnvironment is enum of config environment.
type MTPConfigEnvironment int32

func (e MTPConfigEnvironment) valid() bool {
	return e == 0 || e == 1
}

// String implements fmt.Stringer.
func (e MTPConfigEnvironment) String() string {
	switch e {
	case 0:
		return "production"
	case 1:
		return "test"
	default:
		return "unknown"
	}
}

// Test denotes that environment is test.
func (e MTPConfigEnvironment) Test() bool {
	return e == 1
}

// MTPConfig is a Telegram Desktop storage structure which stores MTProto config info.
//
// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/mtproto/mtproto_config.h.
type MTPConfig struct {
	Environment              MTPConfigEnvironment
	DCOptions                MTPDCOptions
	ChatSizeMax              int32  // default: 200
	MegagroupSizeMax         int32  // default: 10000
	ForwardedCountMax        int32  // default: 100
	OnlineUpdatePeriod       int32  // default: 120000
	OfflineBlurTimeout       int32  // default: 5000
	OfflineIdleTimeout       int32  // default: 30000
	OnlineFocusTimeout       int32  // default: 1000
	OnlineCloudTimeout       int32  // default: 300000
	NotifyCloudDelay         int32  // default: 30000
	NotifyDefaultDelay       int32  // default: 1500
	SavedGifsLimit           int32  // default: 200
	EditTimeLimit            int32  // default: 172800
	RevokeTimeLimit          int32  // default: 172800
	RevokePrivateTimeLimit   int32  // default: 172800
	RevokePrivateInbox       bool   // default: false
	StickersRecentLimit      int32  // default: 30
	StickersFavedLimit       int32  // default: 5
	PinnedDialogsCountMax    int32  // default: 5
	PinnedDialogsInFolderMax int32  // default: 100
	InternalLinksDomain      string // default: "https://t.me/"
	ChannelsReadMediaPeriod  int32  // default: 86400 * 7
	CallReceiveTimeoutMs     int32  // default: 20000
	CallRingTimeoutMs        int32  // default: 90000
	CallConnectTimeoutMs     int32  // default: 30000
	CallPacketTimeoutMs      int32  // default: 10000
	WebFileDCID              int32  // default: 4
	TxtDomainString          string // default: ""
	PhoneCallsEnabled        bool   // default: true
	BlockedMode              bool   // default: false
	CaptionLengthMax         int32  // default: 1024
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/storage_account.cpp#L938.
func readMTPConfig(tgf *tdesktopFile, localKey crypto.Key) (MTPConfig, error) {
	encrypted, err := tgf.readArray()
	if err != nil {
		return MTPConfig{}, errors.Wrap(err, "read encrypted data")
	}

	decrypted, err := decryptLocal(encrypted, localKey)
	if err != nil {
		return MTPConfig{}, errors.Wrap(err, "decrypt data")
	}
	// Skip decrypted data length (uint32).
	decrypted = decrypted[4:]
	root := qtReader{buf: bin.Buffer{Buf: decrypted}}

	cfgReader, err := root.subArray()
	if err != nil {
		return MTPConfig{}, errors.Wrap(err, "read config array")
	}

	var m MTPConfig
	if err := m.deserialize(&cfgReader); err != nil {
		return MTPConfig{}, errors.Wrap(err, "deserialize MTPConfig")
	}
	return m, err
}

func (m *MTPConfig) deserialize(r *qtReader) error {
	version, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read version")
	}
	if version != kVersion {
		return errors.Errorf("wrong version (expected %d, got %d)", kVersion, version)
	}

	environment, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read environment")
	}
	m.Environment = MTPConfigEnvironment(environment)
	if !m.Environment.valid() {
		return errors.Errorf("invalid environment %d", environment)
	}

	{
		sub, err := r.subArray()
		if err != nil {
			return err
		}
		if err := m.DCOptions.deserialize(&sub); err != nil {
			return errors.Wrap(err, "read DC options")
		}
	}

	chatSizeMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read chatSizeMax")
	}
	m.ChatSizeMax = chatSizeMax

	megagroupSizeMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read megagroupSizeMax")
	}
	m.MegagroupSizeMax = megagroupSizeMax

	forwardedCountMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read forwardedCountMax")
	}
	m.ForwardedCountMax = forwardedCountMax

	onlineUpdatePeriod, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read onlineUpdatePeriod")
	}
	m.OnlineUpdatePeriod = onlineUpdatePeriod

	offlineBlurTimeout, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read offlineBlurTimeout")
	}
	m.OfflineBlurTimeout = offlineBlurTimeout

	offlineIdleTimeout, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read offlineIdleTimeout")
	}
	m.OfflineIdleTimeout = offlineIdleTimeout

	onlineFocusTimeout, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read onlineFocusTimeout")
	}
	m.OnlineFocusTimeout = onlineFocusTimeout

	onlineCloudTimeout, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read onlineCloudTimeout")
	}
	m.OnlineCloudTimeout = onlineCloudTimeout

	notifyCloudDelay, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read notifyCloudDelay")
	}
	m.NotifyCloudDelay = notifyCloudDelay

	notifyDefaultDelay, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read notifyDefaultDelay")
	}
	m.NotifyDefaultDelay = notifyDefaultDelay

	savedGifsLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read savedGifsLimit")
	}
	m.SavedGifsLimit = savedGifsLimit

	editTimeLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read editTimeLimit")
	}
	m.EditTimeLimit = editTimeLimit

	revokeTimeLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read revokeTimeLimit")
	}
	m.RevokeTimeLimit = revokeTimeLimit

	revokePrivateTimeLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read revokePrivateTimeLimit")
	}
	m.RevokePrivateTimeLimit = revokePrivateTimeLimit

	revokePrivateInbox, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read revokePrivateInbox")
	}
	m.RevokePrivateInbox = revokePrivateInbox == 1

	stickersRecentLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read stickersRecentLimit")
	}
	m.StickersRecentLimit = stickersRecentLimit

	stickersFavedLimit, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read stickersFavedLimit")
	}
	m.StickersFavedLimit = stickersFavedLimit

	pinnedDialogsCountMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read pinnedDialogsCountMax")
	}
	m.PinnedDialogsCountMax = pinnedDialogsCountMax

	pinnedDialogsInFolderMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read pinnedDialogsInFolderMax")
	}
	m.PinnedDialogsInFolderMax = pinnedDialogsInFolderMax

	internalLinksDomain, err := r.readString()
	if err != nil {
		return errors.Wrap(err, "read internalLinksDomain")
	}
	m.InternalLinksDomain = internalLinksDomain

	channelsReadMediaPeriod, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read channelsReadMediaPeriod")
	}
	m.ChannelsReadMediaPeriod = channelsReadMediaPeriod

	callReceiveTimeoutMs, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read callReceiveTimeoutMs")
	}
	m.CallReceiveTimeoutMs = callReceiveTimeoutMs

	callRingTimeoutMs, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read callRingTimeoutMs")
	}
	m.CallRingTimeoutMs = callRingTimeoutMs

	callConnectTimeoutMs, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read callConnectTimeoutMs")
	}
	m.CallConnectTimeoutMs = callConnectTimeoutMs

	callPacketTimeoutMs, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read callPacketTimeoutMs")
	}
	m.CallPacketTimeoutMs = callPacketTimeoutMs

	webFileDCID, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read webFileDCID")
	}
	m.WebFileDCID = webFileDCID

	txtDomainString, err := r.readString()
	if err != nil {
		return errors.Wrap(err, "read txtDomainString")
	}
	m.TxtDomainString = txtDomainString

	phoneCallsEnabled, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read phoneCallsEnabled")
	}
	m.PhoneCallsEnabled = phoneCallsEnabled == 1

	blockedMode, err := r.readUint32()
	if err != nil {
		return errors.Wrap(err, "read blockedMode")
	}
	m.BlockedMode = blockedMode == 1

	captionLengthMax, err := r.readInt32()
	if err != nil {
		return errors.Wrap(err, "read captionLengthMax")
	}
	m.CaptionLengthMax = captionLengthMax

	return nil
}
