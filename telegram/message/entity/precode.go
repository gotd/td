package entity

import (
	"fmt"

	"github.com/gotd/td/tg"
)

func equalRange(a, b tg.MessageEntityClass) bool {
	return a.GetLength() == b.GetLength() && a.GetOffset() == b.GetOffset()
}

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
			} else {
				fmt.Println("filter", val)
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
