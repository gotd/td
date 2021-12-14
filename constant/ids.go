package constant

const (
	// MaxTDLibChatID is maximum chat TDLib ID.
	MaxTDLibChatID = 999999999999
	// MaxTDLibChannelID is maximum channel TDLib ID.
	MaxTDLibChannelID = 1000000000000 - int64(1<<31)
	// ZeroTDLibChannelID is minimum channel TDLib ID.
	ZeroTDLibChannelID = -1000000000000
	// ZeroTDLibSecretChatID is minimum secret chat TDLib ID.
	ZeroTDLibSecretChatID = -2000000000000
	// MaxTDLibUserID is maximum user TDLib ID.
	MaxTDLibUserID = (1 << 40) - 1
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

// Channel sets TDLibPeerID value as channel.
func (id *TDLibPeerID) Channel(p int64) {
	*id = TDLibPeerID(ZeroTDLibChannelID + (p * -1))
}

// ToPlain converts TDLib ID to plain ID.
func (id TDLibPeerID) ToPlain() (r int64) {
	switch {
	case id.IsUser():
		r = int64(id)
	case id.IsChat():
		r = int64(-id)
	case id.IsChannel():
		r = int64(id) - ZeroTDLibChannelID
		r = -r
	}
	return r
}

// IsUser whether that given ID is user ID.
func (id TDLibPeerID) IsUser() bool {
	return id > 0 && id <= MaxTDLibUserID
}

// IsChat whether that given ID is chat ID.
func (id TDLibPeerID) IsChat() bool {
	return id < 0 && -MaxTDLibChatID <= id
}

// IsChannel whether that given ID is channel ID.
func (id TDLibPeerID) IsChannel() bool {
	return id < 0 &&
		id != ZeroTDLibChannelID &&
		!id.IsChat() &&
		ZeroTDLibChannelID-TDLibPeerID(MaxTDLibChannelID) <= id
}
