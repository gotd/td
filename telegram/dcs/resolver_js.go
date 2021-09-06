//go:build js && wasm
// +build js,wasm

package dcs

// DefaultResolver returns default DC resolver for current platform.
func DefaultResolver() Resolver {
	return Websocket(WebsocketOptions{})
}
