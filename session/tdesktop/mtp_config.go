package tdesktop

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
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
// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/mtproto/mtproto_config.h
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
	WebFileDcId              int32  // default: 4
	TxtDomainString          string // default: ""
	PhoneCallsEnabled        bool   // default: true
	BlockedMode              bool   // default: false
	CaptionLengthMax         int32  // default: 1024
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/storage_account.cpp#L938.
func readMTPConfig(tgf *tdesktopFile, localKey crypto.Key) (MTPConfig, error) {
	encrypted, err := tgf.readArray()
	if err != nil {
		return MTPConfig{}, xerrors.Errorf("read encrypted data: %w", err)
	}

	decrypted, err := decryptLocal(encrypted, localKey)
	if err != nil {
		return MTPConfig{}, xerrors.Errorf("decrypt data: %w", err)
	}
	// Skip decrypted data length (uint32).
	decrypted = decrypted[8:]
	r := qtReader{buf: bin.Buffer{Buf: decrypted}}

	var m MTPConfig
	if err := m.deserialize(&r); err != nil {
		return MTPConfig{}, xerrors.Errorf("deserialize MTPConfig: %w", err)
	}
	return m, err
}

func (m *MTPConfig) deserialize(r *qtReader) error {
	version, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read version: %w", err)
	}
	if version != kVersion {
		return xerrors.Errorf("wrong version (expected %d, got %d)", kVersion, version)
	}

	environment, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read environment: %w", err)
	}
	m.Environment = MTPConfigEnvironment(environment)
	if !m.Environment.valid() {
		return xerrors.Errorf("invalid environment %d", environment)
	}

	{
		sub, err := r.subArray()
		if err != nil {
			return err
		}
		if err := m.DCOptions.deserialize(&sub); err != nil {
			return xerrors.Errorf("read DC options: %w", err)
		}
	}

	chatSizeMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read chatSizeMax: %w", err)
	}
	m.ChatSizeMax = chatSizeMax

	megagroupSizeMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read megagroupSizeMax: %w", err)
	}
	m.MegagroupSizeMax = megagroupSizeMax

	forwardedCountMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read forwardedCountMax: %w", err)
	}
	m.ForwardedCountMax = forwardedCountMax

	onlineUpdatePeriod, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read onlineUpdatePeriod: %w", err)
	}
	m.OnlineUpdatePeriod = onlineUpdatePeriod

	offlineBlurTimeout, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read offlineBlurTimeout: %w", err)
	}
	m.OfflineBlurTimeout = offlineBlurTimeout

	offlineIdleTimeout, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read offlineIdleTimeout: %w", err)
	}
	m.OfflineIdleTimeout = offlineIdleTimeout

	onlineFocusTimeout, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read onlineFocusTimeout: %w", err)
	}
	m.OnlineFocusTimeout = onlineFocusTimeout

	onlineCloudTimeout, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read onlineCloudTimeout: %w", err)
	}
	m.OnlineCloudTimeout = onlineCloudTimeout

	notifyCloudDelay, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read notifyCloudDelay: %w", err)
	}
	m.NotifyCloudDelay = notifyCloudDelay

	notifyDefaultDelay, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read notifyDefaultDelay: %w", err)
	}
	m.NotifyDefaultDelay = notifyDefaultDelay

	savedGifsLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read savedGifsLimit: %w", err)
	}
	m.SavedGifsLimit = savedGifsLimit

	editTimeLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read editTimeLimit: %w", err)
	}
	m.EditTimeLimit = editTimeLimit

	revokeTimeLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read revokeTimeLimit: %w", err)
	}
	m.RevokeTimeLimit = revokeTimeLimit

	revokePrivateTimeLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read revokePrivateTimeLimit: %w", err)
	}
	m.RevokePrivateTimeLimit = revokePrivateTimeLimit

	revokePrivateInbox, err := r.readUint32()
	if err != nil {
		return xerrors.Errorf("read revokePrivateInbox: %w", err)
	}
	m.RevokePrivateInbox = revokePrivateInbox == 1

	stickersRecentLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read stickersRecentLimit: %w", err)
	}
	m.StickersRecentLimit = stickersRecentLimit

	stickersFavedLimit, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read stickersFavedLimit: %w", err)
	}
	m.StickersFavedLimit = stickersFavedLimit

	pinnedDialogsCountMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read pinnedDialogsCountMax: %w", err)
	}
	m.PinnedDialogsCountMax = pinnedDialogsCountMax

	pinnedDialogsInFolderMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read pinnedDialogsInFolderMax: %w", err)
	}
	m.PinnedDialogsInFolderMax = pinnedDialogsInFolderMax

	internalLinksDomain, err := r.readString()
	if err != nil {
		return xerrors.Errorf("read internalLinksDomain: %w", err)
	}
	m.InternalLinksDomain = internalLinksDomain

	channelsReadMediaPeriod, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read channelsReadMediaPeriod: %w", err)
	}
	m.ChannelsReadMediaPeriod = channelsReadMediaPeriod

	callReceiveTimeoutMs, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read callReceiveTimeoutMs: %w", err)
	}
	m.CallReceiveTimeoutMs = callReceiveTimeoutMs

	callRingTimeoutMs, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read callRingTimeoutMs: %w", err)
	}
	m.CallRingTimeoutMs = callRingTimeoutMs

	callConnectTimeoutMs, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read callConnectTimeoutMs: %w", err)
	}
	m.CallConnectTimeoutMs = callConnectTimeoutMs

	callPacketTimeoutMs, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read callPacketTimeoutMs: %w", err)
	}
	m.CallPacketTimeoutMs = callPacketTimeoutMs

	webFileDcId, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read webFileDcId: %w", err)
	}
	m.WebFileDcId = webFileDcId

	txtDomainString, err := r.readString()
	if err != nil {
		return xerrors.Errorf("read txtDomainString: %w", err)
	}
	m.TxtDomainString = txtDomainString

	phoneCallsEnabled, err := r.readUint32()
	if err != nil {
		return xerrors.Errorf("read phoneCallsEnabled: %w", err)
	}
	m.PhoneCallsEnabled = phoneCallsEnabled == 1

	blockedMode, err := r.readUint32()
	if err != nil {
		return xerrors.Errorf("read blockedMode: %w", err)
	}
	m.BlockedMode = blockedMode == 1

	captionLengthMax, err := r.readInt32()
	if err != nil {
		return xerrors.Errorf("read captionLengthMax: %w", err)
	}
	m.CaptionLengthMax = captionLengthMax

	return nil
}
