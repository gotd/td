package members

//go:generate go run -modfile=../../../_tools/go.mod golang.org/x/tools/cmd/stringer -type=Status

// Status defines participant status.
type Status int

const (
	// Plain is status for plain participant.
	Plain Status = iota
	// Creator is status for chat/channel creator.
	Creator
	// Admin is status for chat/channel admin.
	Admin
	// Banned is status for banned user.
	Banned
	// Left is status for user that left chat/channel.
	Left
)
