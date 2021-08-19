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
)

// AccountResetWallPapersRequest represents TL type `account.resetWallPapers#bb3b9804`.
// Delete installed wallpapers
//
// See https://core.telegram.org/method/account.resetWallPapers for reference.
type AccountResetWallPapersRequest struct {
}

// AccountResetWallPapersRequestTypeID is TL type id of AccountResetWallPapersRequest.
const AccountResetWallPapersRequestTypeID = 0xbb3b9804

func (r *AccountResetWallPapersRequest) Zero() bool {
	if r == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (r *AccountResetWallPapersRequest) String() string {
	if r == nil {
		return "AccountResetWallPapersRequest(nil)"
	}
	type Alias AccountResetWallPapersRequest
	return fmt.Sprintf("AccountResetWallPapersRequest%+v", Alias(*r))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountResetWallPapersRequest) TypeID() uint32 {
	return AccountResetWallPapersRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountResetWallPapersRequest) TypeName() string {
	return "account.resetWallPapers"
}

// TypeInfo returns info about TL type.
func (r *AccountResetWallPapersRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.resetWallPapers",
		ID:   AccountResetWallPapersRequestTypeID,
	}
	if r == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (r *AccountResetWallPapersRequest) Encode(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.resetWallPapers#bb3b9804",
		}
	}
	b.PutID(AccountResetWallPapersRequestTypeID)
	return r.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (r *AccountResetWallPapersRequest) EncodeBare(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.resetWallPapers#bb3b9804",
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (r *AccountResetWallPapersRequest) Decode(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.resetWallPapers#bb3b9804",
		}
	}
	if err := b.ConsumeID(AccountResetWallPapersRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "account.resetWallPapers#bb3b9804",
			Underlying: err,
		}
	}
	return r.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (r *AccountResetWallPapersRequest) DecodeBare(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.resetWallPapers#bb3b9804",
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for AccountResetWallPapersRequest.
var (
	_ bin.Encoder     = &AccountResetWallPapersRequest{}
	_ bin.Decoder     = &AccountResetWallPapersRequest{}
	_ bin.BareEncoder = &AccountResetWallPapersRequest{}
	_ bin.BareDecoder = &AccountResetWallPapersRequest{}
)

// AccountResetWallPapers invokes method account.resetWallPapers#bb3b9804 returning error if any.
// Delete installed wallpapers
//
// See https://core.telegram.org/method/account.resetWallPapers for reference.
func (c *Client) AccountResetWallPapers(ctx context.Context) (bool, error) {
	var result BoolBox

	request := &AccountResetWallPapersRequest{}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
