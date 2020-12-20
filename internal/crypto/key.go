package crypto

// AuthKeyWithID is a AuthKey with cached id.
type AuthKeyWithID struct {
	AuthKey   AuthKey
	AuthKeyID [8]byte
}

// Zero reports whether AuthKey is zero value.
func (a AuthKeyWithID) Zero() bool {
	return a == AuthKeyWithID{}
}
