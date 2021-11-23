package constant

const (
	// MaxChatID is maximum chat ID.
	MaxChatID = 999999999999
	// MaxChannelID is maximum channel ID.
	MaxChannelID = 1000000000000 - int64(1<<31)
	// ZeroChannelID is minimum channel ID.
	ZeroChannelID = -1000000000000
	// ZeroSecretChatID is minimum secret chat ID.
	ZeroSecretChatID = -2000000000000
	// MaxUserID is maximum user ID.
	MaxUserID = (1 << 40) - 1
)
