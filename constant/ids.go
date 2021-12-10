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

// IsUserTDLibID whether that given ID is user ID.
func IsUserTDLibID(id int64) bool {
	return id > 0 && id <= MaxTDLibUserID
}

// IsChatTDLibID whether that given ID is chat ID.
func IsChatTDLibID(id int64) bool {
	return id < 0 && -MaxTDLibChatID <= id
}

// IsChannelTDLibID whether that given ID is channel ID.
func IsChannelTDLibID(id int64) bool {
	return id < 0 &&
		id != ZeroTDLibChannelID &&
		!IsChatTDLibID(id) &&
		ZeroTDLibChannelID-MaxTDLibChannelID <= id

}
