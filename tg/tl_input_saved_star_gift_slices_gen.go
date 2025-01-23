//go:build !no_gotd_slices
// +build !no_gotd_slices

// Code generated by gotdgen, DO NOT EDIT.

package tg

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdjson"
	"github.com/gotd/td/tdp"
	"github.com/gotd/td/tgerr"
)

// No-op definition for keeping imports.
var (
	_ = bin.Buffer{}
	_ = context.Background()
	_ = fmt.Stringer(nil)
	_ = strings.Builder{}
	_ = errors.Is
	_ = multierr.AppendInto
	_ = sort.Ints
	_ = tdp.Format
	_ = tgerr.Error{}
	_ = tdjson.Encoder{}
)

// InputSavedStarGiftClassArray is adapter for slice of InputSavedStarGiftClass.
type InputSavedStarGiftClassArray []InputSavedStarGiftClass

// Sort sorts slice of InputSavedStarGiftClass.
func (s InputSavedStarGiftClassArray) Sort(less func(a, b InputSavedStarGiftClass) bool) InputSavedStarGiftClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of InputSavedStarGiftClass.
func (s InputSavedStarGiftClassArray) SortStable(less func(a, b InputSavedStarGiftClass) bool) InputSavedStarGiftClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of InputSavedStarGiftClass.
func (s InputSavedStarGiftClassArray) Retain(keep func(x InputSavedStarGiftClass) bool) InputSavedStarGiftClassArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s InputSavedStarGiftClassArray) First() (v InputSavedStarGiftClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s InputSavedStarGiftClassArray) Last() (v InputSavedStarGiftClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftClassArray) PopFirst() (v InputSavedStarGiftClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero InputSavedStarGiftClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftClassArray) Pop() (v InputSavedStarGiftClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsInputSavedStarGiftUser returns copy with only InputSavedStarGiftUser constructors.
func (s InputSavedStarGiftClassArray) AsInputSavedStarGiftUser() (to InputSavedStarGiftUserArray) {
	for _, elem := range s {
		value, ok := elem.(*InputSavedStarGiftUser)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsInputSavedStarGiftChat returns copy with only InputSavedStarGiftChat constructors.
func (s InputSavedStarGiftClassArray) AsInputSavedStarGiftChat() (to InputSavedStarGiftChatArray) {
	for _, elem := range s {
		value, ok := elem.(*InputSavedStarGiftChat)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// InputSavedStarGiftUserArray is adapter for slice of InputSavedStarGiftUser.
type InputSavedStarGiftUserArray []InputSavedStarGiftUser

// Sort sorts slice of InputSavedStarGiftUser.
func (s InputSavedStarGiftUserArray) Sort(less func(a, b InputSavedStarGiftUser) bool) InputSavedStarGiftUserArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of InputSavedStarGiftUser.
func (s InputSavedStarGiftUserArray) SortStable(less func(a, b InputSavedStarGiftUser) bool) InputSavedStarGiftUserArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of InputSavedStarGiftUser.
func (s InputSavedStarGiftUserArray) Retain(keep func(x InputSavedStarGiftUser) bool) InputSavedStarGiftUserArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s InputSavedStarGiftUserArray) First() (v InputSavedStarGiftUser, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s InputSavedStarGiftUserArray) Last() (v InputSavedStarGiftUser, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftUserArray) PopFirst() (v InputSavedStarGiftUser, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero InputSavedStarGiftUser
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftUserArray) Pop() (v InputSavedStarGiftUser, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// InputSavedStarGiftChatArray is adapter for slice of InputSavedStarGiftChat.
type InputSavedStarGiftChatArray []InputSavedStarGiftChat

// Sort sorts slice of InputSavedStarGiftChat.
func (s InputSavedStarGiftChatArray) Sort(less func(a, b InputSavedStarGiftChat) bool) InputSavedStarGiftChatArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of InputSavedStarGiftChat.
func (s InputSavedStarGiftChatArray) SortStable(less func(a, b InputSavedStarGiftChat) bool) InputSavedStarGiftChatArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of InputSavedStarGiftChat.
func (s InputSavedStarGiftChatArray) Retain(keep func(x InputSavedStarGiftChat) bool) InputSavedStarGiftChatArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s InputSavedStarGiftChatArray) First() (v InputSavedStarGiftChat, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s InputSavedStarGiftChatArray) Last() (v InputSavedStarGiftChat, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftChatArray) PopFirst() (v InputSavedStarGiftChat, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero InputSavedStarGiftChat
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *InputSavedStarGiftChatArray) Pop() (v InputSavedStarGiftChat, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
