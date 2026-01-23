package constant

const (
	// MaxTDLibChatID is maximum chat TDLib ID.
	MaxTDLibChatID = 999999999999
	// MaxTDLibChannelID is maximum channel TDLib ID (non-monoforum).
	MaxTDLibChannelID = 1000000000000 - int64(1<<31) // 997852516352
	// ZeroTDLibChannelID is base for channel IDs.
	ZeroTDLibChannelID = -1000000000000
	// ZeroTDLibSecretChatID is minimum secret chat TDLib ID.
	ZeroTDLibSecretChatID = -2000000000000
	// MaxTDLibUserID is maximum user TDLib ID.
	MaxTDLibUserID = (1 << 40) - 1

	// MinMTProtoMonoforumID is the minimum MTProto ID for monoforums.
	MinMTProtoMonoforumID = 1002147483649
	// MaxMTProtoMonoforumID is the maximum MTProto ID for monoforums.
	MaxMTProtoMonoforumID = 3000000000000

	// MinBotAPIMonoforumID = -(1e12 + MaxMTProtoMonoforumID)
	MinBotAPIMonoforumID = -4000000000000
	// MaxBotAPIMonoforumID = -(1e12 + MinMTProtoMonoforumID)
	MaxBotAPIMonoforumID = -2002147483649
)

// TDLibPeerID is TDLib's peer ID.
type TDLibPeerID int64

// User sets TDLibPeerID value as user.
func (id *TDLibPeerID) User(p int64) {
	*id = TDLibPeerID(p)
}

// Chat sets TDLibPeerID value as chat.
func (id *TDLibPeerID) Chat(p int64) {
	*id = TDLibPeerID(-p)
}

// Channel sets TDLibPeerID value as channel (including monoforum).
func (id *TDLibPeerID) Channel(p int64) {
	*id = TDLibPeerID(ZeroTDLibChannelID - p) // same as ZeroTDLibChannelID + (p * -1)
}

// ToPlain converts TDLib ID to plain (MTProto) ID.
func (id TDLibPeerID) ToPlain() (r int64) {
	switch {
	case id.IsUser():
		r = int64(id)
	case id.IsChat():
		r = -int64(id)
	case id.IsChannel():
		r = -(int64(id) - ZeroTDLibChannelID)
	}
	return r
}

// IsUser whether that given ID is user ID.
func (id TDLibPeerID) IsUser() bool {
	return id > 0 && id <= MaxTDLibUserID
}

// IsChat whether that given ID is chat ID.
func (id TDLibPeerID) IsChat() bool {
	return id < 0 && id >= -MaxTDLibChatID
}

// IsMonoforum checks if the ID belongs to a monoforum.
func (id TDLibPeerID) IsMonoforum() bool {
	return id >= MinBotAPIMonoforumID && id <= MaxBotAPIMonoforumID
}

// IsChannel whether that given ID is channel ID (including monoforums).
func (id TDLibPeerID) IsChannel() bool {
	if id >= 0 {
		return false
	}
	if id == ZeroTDLibChannelID || id.IsChat() {
		return false
	}
	// Regular channels: -1997852516352 to -1000000000001
	if id >= -1997852516352 && id <= -1000000000001 {
		return true
	}
	// Monoforums: -4000000000000 to -2002147483649
	return id >= MinBotAPIMonoforumID && id <= MaxBotAPIMonoforumID
}
