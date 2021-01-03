package telegram

import (
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
)

// SessionStorage is alias of mtproto.SessionStorage.
//
// Deprecated.
type SessionStorage = session.Storage

// FileSessionStorage is alias of mtproto.FileSessionStorage.
//
// Deprecated.
type FileSessionStorage = session.FileStorage

// Error represents RPC error returned to request.
//
// Deprecated.
type Error = mtproto.Error
