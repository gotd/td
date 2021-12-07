package entity

import "github.com/gotd/td/tg"

// UserResolver is callback for resolving InputUser by ID.
type UserResolver = func(id int64) (tg.InputUserClass, error)
