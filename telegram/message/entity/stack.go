package entity

type stackElem struct {
	offset     int
	utf8offset int
	tag        string
	format     Formatter
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
