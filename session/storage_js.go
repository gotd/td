//go:build js && wasm
// +build js,wasm

package session

import (
	"context"
	"syscall/js"

	"github.com/go-faster/errors"
)

// WebLocalStorage is a Web Storage API based session storage.
type WebLocalStorage struct {
	Key string
}

func getStorage() (js.Value, bool) {
	localStorage := js.Global().Get("localStorage")

	if localStorage.IsUndefined() || localStorage.IsNull() {
		return js.Value{}, false
	}

	const testValue = "__test__"
	localStorage.Set(testValue, testValue)
	value := localStorage.Get(testValue)
	if value.IsUndefined() || value.IsNull() {
		return js.Value{}, false
	}
	localStorage.Delete(testValue)

	return localStorage, true
}

// ErrLocalStorageIsNotAvailable is returned if localStorage is not available and Storage can't use it.
var ErrLocalStorageIsNotAvailable = errors.New("localStorage is not available")

func catch(err *error) { // nolint:gocritic
	if r := recover(); r != nil {
		rErr, ok := r.(error)
		if !ok {
			*err = errors.Errorf("catch: %v", r)
		} else {
			*err = errors.Wrap(rErr, "catch")
		}
	}
}

// LoadSession loads session using Web Storage API.
func (w WebLocalStorage) LoadSession(_ context.Context) (_ []byte, rerr error) {
	defer catch(&rerr)

	if w.Key == "" {
		return nil, errors.Errorf("invalid key %q", w.Key)
	}

	store, ok := getStorage()
	if !ok {
		return nil, ErrLocalStorageIsNotAvailable
	}

	value := store.Call("getItem", w.Key)
	if value.IsNull() || value.IsUndefined() {
		return nil, ErrNotFound
	}

	return []byte(value.String()), nil
}

// StoreSession saves session using Web Storage API.
func (w WebLocalStorage) StoreSession(_ context.Context, data []byte) (rerr error) {
	defer catch(&rerr)

	if w.Key == "" {
		return errors.Errorf("invalid key %q", w.Key)
	}

	store, ok := getStorage()
	if !ok {
		return ErrLocalStorageIsNotAvailable
	}

	store.Call("setItem", w.Key, string(data))
	return nil
}
