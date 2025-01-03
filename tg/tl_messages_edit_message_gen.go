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

// MessagesEditMessageRequest represents TL type `messages.editMessage#dfd14005`.
// Edit message
//
// See https://core.telegram.org/method/messages.editMessage for reference.
type MessagesEditMessageRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Disable webpage preview
	NoWebpage bool
	// If set, any eventual webpage preview will be shown on top of the message instead of at
	// the bottom.
	InvertMedia bool
	// Where was the message sent
	Peer InputPeerClass
	// ID of the message to edit
	ID int
	// New message
	//
	// Use SetMessage and GetMessage helpers.
	Message string
	// New attached media
	//
	// Use SetMedia and GetMedia helpers.
	Media InputMediaClass
	// Reply markup for inline keyboards
	//
	// Use SetReplyMarkup and GetReplyMarkup helpers.
	ReplyMarkup ReplyMarkupClass
	// Message entities for styled text¹
	//
	// Links:
	//  1) https://core.telegram.org/api/entities
	//
	// Use SetEntities and GetEntities helpers.
	Entities []MessageEntityClass
	// Scheduled message date for scheduled messages¹
	//
	// Links:
	//  1) https://core.telegram.org/api/scheduled-messages
	//
	// Use SetScheduleDate and GetScheduleDate helpers.
	ScheduleDate int
	// If specified, edits a quick reply shortcut message, instead »¹.
	//
	// Links:
	//  1) https://core.telegram.org/api/business#quick-reply-shortcuts
	//
	// Use SetQuickReplyShortcutID and GetQuickReplyShortcutID helpers.
	QuickReplyShortcutID int
}

// MessagesEditMessageRequestTypeID is TL type id of MessagesEditMessageRequest.
const MessagesEditMessageRequestTypeID = 0xdfd14005

// Ensuring interfaces in compile-time for MessagesEditMessageRequest.
var (
	_ bin.Encoder     = &MessagesEditMessageRequest{}
	_ bin.Decoder     = &MessagesEditMessageRequest{}
	_ bin.BareEncoder = &MessagesEditMessageRequest{}
	_ bin.BareDecoder = &MessagesEditMessageRequest{}
)

func (e *MessagesEditMessageRequest) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Flags.Zero()) {
		return false
	}
	if !(e.NoWebpage == false) {
		return false
	}
	if !(e.InvertMedia == false) {
		return false
	}
	if !(e.Peer == nil) {
		return false
	}
	if !(e.ID == 0) {
		return false
	}
	if !(e.Message == "") {
		return false
	}
	if !(e.Media == nil) {
		return false
	}
	if !(e.ReplyMarkup == nil) {
		return false
	}
	if !(e.Entities == nil) {
		return false
	}
	if !(e.ScheduleDate == 0) {
		return false
	}
	if !(e.QuickReplyShortcutID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *MessagesEditMessageRequest) String() string {
	if e == nil {
		return "MessagesEditMessageRequest(nil)"
	}
	type Alias MessagesEditMessageRequest
	return fmt.Sprintf("MessagesEditMessageRequest%+v", Alias(*e))
}

// FillFrom fills MessagesEditMessageRequest from given interface.
func (e *MessagesEditMessageRequest) FillFrom(from interface {
	GetNoWebpage() (value bool)
	GetInvertMedia() (value bool)
	GetPeer() (value InputPeerClass)
	GetID() (value int)
	GetMessage() (value string, ok bool)
	GetMedia() (value InputMediaClass, ok bool)
	GetReplyMarkup() (value ReplyMarkupClass, ok bool)
	GetEntities() (value []MessageEntityClass, ok bool)
	GetScheduleDate() (value int, ok bool)
	GetQuickReplyShortcutID() (value int, ok bool)
}) {
	e.NoWebpage = from.GetNoWebpage()
	e.InvertMedia = from.GetInvertMedia()
	e.Peer = from.GetPeer()
	e.ID = from.GetID()
	if val, ok := from.GetMessage(); ok {
		e.Message = val
	}

	if val, ok := from.GetMedia(); ok {
		e.Media = val
	}

	if val, ok := from.GetReplyMarkup(); ok {
		e.ReplyMarkup = val
	}

	if val, ok := from.GetEntities(); ok {
		e.Entities = val
	}

	if val, ok := from.GetScheduleDate(); ok {
		e.ScheduleDate = val
	}

	if val, ok := from.GetQuickReplyShortcutID(); ok {
		e.QuickReplyShortcutID = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesEditMessageRequest) TypeID() uint32 {
	return MessagesEditMessageRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesEditMessageRequest) TypeName() string {
	return "messages.editMessage"
}

// TypeInfo returns info about TL type.
func (e *MessagesEditMessageRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.editMessage",
		ID:   MessagesEditMessageRequestTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "NoWebpage",
			SchemaName: "no_webpage",
			Null:       !e.Flags.Has(1),
		},
		{
			Name:       "InvertMedia",
			SchemaName: "invert_media",
			Null:       !e.Flags.Has(16),
		},
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "ID",
			SchemaName: "id",
		},
		{
			Name:       "Message",
			SchemaName: "message",
			Null:       !e.Flags.Has(11),
		},
		{
			Name:       "Media",
			SchemaName: "media",
			Null:       !e.Flags.Has(14),
		},
		{
			Name:       "ReplyMarkup",
			SchemaName: "reply_markup",
			Null:       !e.Flags.Has(2),
		},
		{
			Name:       "Entities",
			SchemaName: "entities",
			Null:       !e.Flags.Has(3),
		},
		{
			Name:       "ScheduleDate",
			SchemaName: "schedule_date",
			Null:       !e.Flags.Has(15),
		},
		{
			Name:       "QuickReplyShortcutID",
			SchemaName: "quick_reply_shortcut_id",
			Null:       !e.Flags.Has(17),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (e *MessagesEditMessageRequest) SetFlags() {
	if !(e.NoWebpage == false) {
		e.Flags.Set(1)
	}
	if !(e.InvertMedia == false) {
		e.Flags.Set(16)
	}
	if !(e.Message == "") {
		e.Flags.Set(11)
	}
	if !(e.Media == nil) {
		e.Flags.Set(14)
	}
	if !(e.ReplyMarkup == nil) {
		e.Flags.Set(2)
	}
	if !(e.Entities == nil) {
		e.Flags.Set(3)
	}
	if !(e.ScheduleDate == 0) {
		e.Flags.Set(15)
	}
	if !(e.QuickReplyShortcutID == 0) {
		e.Flags.Set(17)
	}
}

// Encode implements bin.Encoder.
func (e *MessagesEditMessageRequest) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode messages.editMessage#dfd14005 as nil")
	}
	b.PutID(MessagesEditMessageRequestTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *MessagesEditMessageRequest) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode messages.editMessage#dfd14005 as nil")
	}
	e.SetFlags()
	if err := e.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field flags: %w", err)
	}
	if e.Peer == nil {
		return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field peer is nil")
	}
	if err := e.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field peer: %w", err)
	}
	b.PutInt(e.ID)
	if e.Flags.Has(11) {
		b.PutString(e.Message)
	}
	if e.Flags.Has(14) {
		if e.Media == nil {
			return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field media is nil")
		}
		if err := e.Media.Encode(b); err != nil {
			return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field media: %w", err)
		}
	}
	if e.Flags.Has(2) {
		if e.ReplyMarkup == nil {
			return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field reply_markup is nil")
		}
		if err := e.ReplyMarkup.Encode(b); err != nil {
			return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field reply_markup: %w", err)
		}
	}
	if e.Flags.Has(3) {
		b.PutVectorHeader(len(e.Entities))
		for idx, v := range e.Entities {
			if v == nil {
				return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field entities element with index %d is nil", idx)
			}
			if err := v.Encode(b); err != nil {
				return fmt.Errorf("unable to encode messages.editMessage#dfd14005: field entities element with index %d: %w", idx, err)
			}
		}
	}
	if e.Flags.Has(15) {
		b.PutInt(e.ScheduleDate)
	}
	if e.Flags.Has(17) {
		b.PutInt(e.QuickReplyShortcutID)
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *MessagesEditMessageRequest) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode messages.editMessage#dfd14005 to nil")
	}
	if err := b.ConsumeID(MessagesEditMessageRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.editMessage#dfd14005: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *MessagesEditMessageRequest) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode messages.editMessage#dfd14005 to nil")
	}
	{
		if err := e.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field flags: %w", err)
		}
	}
	e.NoWebpage = e.Flags.Has(1)
	e.InvertMedia = e.Flags.Has(16)
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field peer: %w", err)
		}
		e.Peer = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field id: %w", err)
		}
		e.ID = value
	}
	if e.Flags.Has(11) {
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field message: %w", err)
		}
		e.Message = value
	}
	if e.Flags.Has(14) {
		value, err := DecodeInputMedia(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field media: %w", err)
		}
		e.Media = value
	}
	if e.Flags.Has(2) {
		value, err := DecodeReplyMarkup(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field reply_markup: %w", err)
		}
		e.ReplyMarkup = value
	}
	if e.Flags.Has(3) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field entities: %w", err)
		}

		if headerLen > 0 {
			e.Entities = make([]MessageEntityClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeMessageEntity(b)
			if err != nil {
				return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field entities: %w", err)
			}
			e.Entities = append(e.Entities, value)
		}
	}
	if e.Flags.Has(15) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field schedule_date: %w", err)
		}
		e.ScheduleDate = value
	}
	if e.Flags.Has(17) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messages.editMessage#dfd14005: field quick_reply_shortcut_id: %w", err)
		}
		e.QuickReplyShortcutID = value
	}
	return nil
}

// SetNoWebpage sets value of NoWebpage conditional field.
func (e *MessagesEditMessageRequest) SetNoWebpage(value bool) {
	if value {
		e.Flags.Set(1)
		e.NoWebpage = true
	} else {
		e.Flags.Unset(1)
		e.NoWebpage = false
	}
}

// GetNoWebpage returns value of NoWebpage conditional field.
func (e *MessagesEditMessageRequest) GetNoWebpage() (value bool) {
	if e == nil {
		return
	}
	return e.Flags.Has(1)
}

// SetInvertMedia sets value of InvertMedia conditional field.
func (e *MessagesEditMessageRequest) SetInvertMedia(value bool) {
	if value {
		e.Flags.Set(16)
		e.InvertMedia = true
	} else {
		e.Flags.Unset(16)
		e.InvertMedia = false
	}
}

// GetInvertMedia returns value of InvertMedia conditional field.
func (e *MessagesEditMessageRequest) GetInvertMedia() (value bool) {
	if e == nil {
		return
	}
	return e.Flags.Has(16)
}

// GetPeer returns value of Peer field.
func (e *MessagesEditMessageRequest) GetPeer() (value InputPeerClass) {
	if e == nil {
		return
	}
	return e.Peer
}

// GetID returns value of ID field.
func (e *MessagesEditMessageRequest) GetID() (value int) {
	if e == nil {
		return
	}
	return e.ID
}

// SetMessage sets value of Message conditional field.
func (e *MessagesEditMessageRequest) SetMessage(value string) {
	e.Flags.Set(11)
	e.Message = value
}

// GetMessage returns value of Message conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetMessage() (value string, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(11) {
		return value, false
	}
	return e.Message, true
}

// SetMedia sets value of Media conditional field.
func (e *MessagesEditMessageRequest) SetMedia(value InputMediaClass) {
	e.Flags.Set(14)
	e.Media = value
}

// GetMedia returns value of Media conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetMedia() (value InputMediaClass, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(14) {
		return value, false
	}
	return e.Media, true
}

// SetReplyMarkup sets value of ReplyMarkup conditional field.
func (e *MessagesEditMessageRequest) SetReplyMarkup(value ReplyMarkupClass) {
	e.Flags.Set(2)
	e.ReplyMarkup = value
}

// GetReplyMarkup returns value of ReplyMarkup conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetReplyMarkup() (value ReplyMarkupClass, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(2) {
		return value, false
	}
	return e.ReplyMarkup, true
}

// SetEntities sets value of Entities conditional field.
func (e *MessagesEditMessageRequest) SetEntities(value []MessageEntityClass) {
	e.Flags.Set(3)
	e.Entities = value
}

// GetEntities returns value of Entities conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetEntities() (value []MessageEntityClass, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(3) {
		return value, false
	}
	return e.Entities, true
}

// SetScheduleDate sets value of ScheduleDate conditional field.
func (e *MessagesEditMessageRequest) SetScheduleDate(value int) {
	e.Flags.Set(15)
	e.ScheduleDate = value
}

// GetScheduleDate returns value of ScheduleDate conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetScheduleDate() (value int, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(15) {
		return value, false
	}
	return e.ScheduleDate, true
}

// SetQuickReplyShortcutID sets value of QuickReplyShortcutID conditional field.
func (e *MessagesEditMessageRequest) SetQuickReplyShortcutID(value int) {
	e.Flags.Set(17)
	e.QuickReplyShortcutID = value
}

// GetQuickReplyShortcutID returns value of QuickReplyShortcutID conditional field and
// boolean which is true if field was set.
func (e *MessagesEditMessageRequest) GetQuickReplyShortcutID() (value int, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(17) {
		return value, false
	}
	return e.QuickReplyShortcutID, true
}

// MapEntities returns field Entities wrapped in MessageEntityClassArray helper.
func (e *MessagesEditMessageRequest) MapEntities() (value MessageEntityClassArray, ok bool) {
	if !e.Flags.Has(3) {
		return value, false
	}
	return MessageEntityClassArray(e.Entities), true
}

// MessagesEditMessage invokes method messages.editMessage#dfd14005 returning error if any.
// Edit message
//
// Possible errors:
//
//	400 BOT_DOMAIN_INVALID: Bot domain invalid.
//	400 BOT_INVALID: This is not a valid bot.
//	400 BUTTON_COPY_TEXT_INVALID: The specified keyboardButtonCopy.copy_text is invalid.
//	400 BUTTON_DATA_INVALID: The data of one or more of the buttons you provided is invalid.
//	400 BUTTON_TYPE_INVALID: The type of one or more of the buttons you provided is invalid.
//	400 BUTTON_URL_INVALID: Button URL invalid.
//	400 CHANNEL_INVALID: The provided channel is invalid.
//	406 CHANNEL_PRIVATE: You haven't joined this channel/supergroup.
//	403 CHAT_ADMIN_REQUIRED: You must be an admin in this chat to do this.
//	400 CHAT_FORWARDS_RESTRICTED: You can't forward messages from a protected chat.
//	403 CHAT_SEND_GIFS_FORBIDDEN: You can't send gifs in this chat.
//	403 CHAT_WRITE_FORBIDDEN: You can't write in this chat.
//	400 DOCUMENT_INVALID: The specified document is invalid.
//	400 ENTITIES_TOO_LONG: You provided too many styled message entities.
//	400 ENTITY_BOUNDS_INVALID: A specified entity offset or length is invalid, see here » for info on how to properly compute the entity offset/length.
//	400 FILE_PARTS_INVALID: The number of file parts is invalid.
//	400 IMAGE_PROCESS_FAILED: Failure while processing image.
//	403 INLINE_BOT_REQUIRED: Only the inline bot can edit message.
//	400 INPUT_USER_DEACTIVATED: The specified user was deleted.
//	400 MEDIA_CAPTION_TOO_LONG: The caption is too long.
//	400 MEDIA_EMPTY: The provided media object is invalid.
//	400 MEDIA_GROUPED_INVALID: You tried to send media of different types in an album.
//	400 MEDIA_INVALID: Media invalid.
//	400 MEDIA_NEW_INVALID: The new media is invalid.
//	400 MEDIA_PREV_INVALID: Previous media invalid.
//	400 MEDIA_TTL_INVALID: The specified media TTL is invalid.
//	403 MESSAGE_AUTHOR_REQUIRED: Message author required.
//	400 MESSAGE_EDIT_TIME_EXPIRED: You can't edit this message anymore, too much time has passed since its creation.
//	400 MESSAGE_EMPTY: The provided message is empty.
//	400 MESSAGE_ID_INVALID: The provided message id is invalid.
//	400 MESSAGE_NOT_MODIFIED: The provided message data is identical to the previous message data, the message wasn't modified.
//	400 MESSAGE_TOO_LONG: The provided message is too long.
//	400 MSG_ID_INVALID: Invalid message ID provided.
//	500 MSG_WAIT_FAILED: A waiting call returned an error.
//	400 PEER_ID_INVALID: The provided peer id is invalid.
//	400 PEER_TYPES_INVALID: The passed keyboardButtonSwitchInline.peer_types field is invalid.
//	400 REPLY_MARKUP_INVALID: The provided reply markup is invalid.
//	400 REPLY_MARKUP_TOO_LONG: The specified reply_markup is too long.
//	400 SCHEDULE_DATE_INVALID: Invalid schedule date provided.
//	400 USER_BANNED_IN_CHANNEL: You're banned from sending messages in supergroups/channels.
//	400 WEBPAGE_NOT_FOUND: A preview for the specified webpage url could not be generated.
//
// See https://core.telegram.org/method/messages.editMessage for reference.
// Can be used by bots.
func (c *Client) MessagesEditMessage(ctx context.Context, request *MessagesEditMessageRequest) (UpdatesClass, error) {
	var result UpdatesBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.Updates, nil
}
