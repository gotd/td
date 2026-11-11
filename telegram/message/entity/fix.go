package entity

import (
	"cmp"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
)

// SortEntities sorts entities as TDLib does it: by offset ascending, then by
// length descending, then by type priority (outer entities first).
//
// The sort is stable, so entities with equal keys keep their original order.
func SortEntities(entity []tg.MessageEntityClass) {
	slices.SortStableFunc(entity, compareEntities)
}

// compareEntities mirrors TDLib MessageEntity::operator<.
func compareEntities(a, b tg.MessageEntityClass) int {
	if c := cmp.Compare(a.GetOffset(), b.GetOffset()); c != 0 {
		return c
	}
	if c := cmp.Compare(b.GetLength(), a.GetLength()); c != 0 {
		return c
	}
	return cmp.Compare(entityPriority(a), entityPriority(b))
}

// entityPriority returns TDLib type priority of entity
// (MessageEntity::get_type_priority). Lower priority goes first.
func entityPriority(e tg.MessageEntityClass) int {
	switch e := e.(type) {
	case *tg.MessageEntityBlockquote:
		return 0
	case *tg.MessageEntityPre:
		if e.Language != "" {
			return 10 // PreCode
		}
		return 11
	case *tg.MessageEntityCode:
		return 20
	case *tg.MessageEntityFormattedDate:
		return 30
	case *tg.MessageEntityTextURL,
		*tg.MessageEntityMentionName,
		*tg.InputMessageEntityMentionName:
		return 49
	case *tg.MessageEntityBold:
		return 90
	case *tg.MessageEntityItalic:
		return 91
	case *tg.MessageEntityUnderline:
		return 92
	case *tg.MessageEntityStrike:
		return 93
	case *tg.MessageEntitySpoiler:
		return 94
	case *tg.MessageEntityCustomEmoji:
		return 99
	default:
		// Mention, Hashtag, BotCommand, URL, Email, Cashtag, Phone,
		// BankCard, TonAddress and unknown entities.
		return 50
	}
}

// setLength sets Length field of entity.
func setLength(index, value int, slice []tg.MessageEntityClass) {
	reflect.ValueOf(&slice[index]).
		Elem().Elem().Elem().
		FieldByName("Length").
		SetInt(int64(value))
}

// fixEntities trims space, if needed and fixes entities offsets.
func (b *Builder) fixEntities(msg string, entities []tg.MessageEntityClass) (string, []tg.MessageEntityClass) {
	// If there are no entities or last text block does not have entities,
	// so we just return built message.
	if len(b.lengths) == 0 || b.lastFormatIndex >= len(entities) {
		return msg, entities
	}

	// Since Telegram client does not handle space after formatted message
	// we should compute length of the last block to trim it.
	// Get first entity of last text block.
	entity := b.lengths[len(b.lengths)-1]
	offset := entity.offset
	length := entity.length
	// Get last text block.
	lastBlock := msg[offset:]
	// Trim this block.
	trimmed := strings.TrimRightFunc(lastBlock, unicode.IsSpace)

	// If there are a difference, we should change length of the all entities.
	if length >= len(lastBlock) && len(trimmed) != len(lastBlock) {
		length := ComputeLength(trimmed)
		for idx := range entities[b.lastFormatIndex:] {
			setLength(idx, length, entities[b.lastFormatIndex:])
		}

		msg = msg[:offset+len(trimmed)]
	}

	return msg, entities
}

// Raw returns raw result and resets builder without fixing spaces.
func (b *Builder) Raw() (string, []tg.MessageEntityClass) {
	msg := b.message.String()
	entities := b.entities
	b.Reset()
	return msg, entities
}

// Complete returns build result and resets builder.
//
// Entities are nil if there are none, so that optional entities vectors of
// requests are omitted instead of sent empty (as Telegram Desktop does).
func (b *Builder) Complete() (string, []tg.MessageEntityClass) {
	msg, entities := b.Raw()
	msg, entities = b.fixEntities(msg, entities)
	if len(entities) == 0 {
		return msg, nil
	}
	SortEntities(entities)
	return msg, entities
}

// ShrinkPreCode merges following <pre> and <code> entities, if needed.
//
// This function is used by formatters to be compliant with TDLib.
func (b *Builder) ShrinkPreCode() {
	b.entities = shrinkPreCode(b.entities)
}

// equalRange compares ranges of given entities.
func equalRange(a, b tg.MessageEntityClass) bool {
	return a.GetLength() == b.GetLength() && a.GetOffset() == b.GetOffset()
}

// shrinkPreCode merges following <pre> and <code> entities, if needed.
func shrinkPreCode(entities []tg.MessageEntityClass) []tg.MessageEntityClass {
	for i, j := 0, len(entities)-1; i < j; i, j = i+1, j-1 {
		entities[i], entities[j] = entities[j], entities[i]
	}

	filter := func(keep func(prev, cur tg.MessageEntityClass) bool) []tg.MessageEntityClass {
		n := 0
		for i, val := range entities {
			if i == 0 || keep(entities[i-1], val) {
				entities[n] = val
				n++
			}
		}
		return entities[:n]
	}

	isPreCode := func(class tg.MessageEntityClass) bool {
		typeID := class.TypeID()
		return typeID == tg.MessageEntityCodeTypeID || typeID == tg.MessageEntityPreTypeID
	}

	hasLang := func(class tg.MessageEntityClass) bool {
		pre, ok := class.(*tg.MessageEntityPre)
		return ok && pre.Language != ""
	}

	resetLang := func(class tg.MessageEntityClass) {
		pre, ok := class.(*tg.MessageEntityPre)
		if !ok {
			return
		}
		pre.Language = ""
	}

	return filter(func(prev, cur tg.MessageEntityClass) bool {
		if !isPreCode(prev) ||
			!isPreCode(cur) ||
			prev.TypeID() == cur.TypeID() {
			// Keep if not is Pre/Code entities or if they are same.
			return true
		}
		if !equalRange(prev, cur) {
			resetLang(prev)
			resetLang(cur)
			return true
		}
		return !hasLang(prev)
	})
}
