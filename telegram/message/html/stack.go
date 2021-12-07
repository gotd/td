package html

import "github.com/gotd/td/telegram/message/entity"

type stackElem struct {
	token  entity.Token
	tag    string
	attr   string
	format entity.Formatter
}

type stack []stackElem

func (s *stack) push(e stackElem) {
	*s = append(*s, e)
}

func (s *stack) last() (stackElem, bool) {
	l := len(*s)
	if l == 0 {
		return stackElem{}, false
	}

	elem := (*s)[l-1]
	return elem, true
}

func (s *stack) pop() (stackElem, bool) {
	e, ok := s.last()
	if !ok {
		return stackElem{}, false
	}
	*s = (*s)[:len(*s)-1]
	return e, true
}
