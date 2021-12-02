package entity

import (
	"sort"

	"github.com/gotd/td/tg"
)

func sortEntities(entity []tg.MessageEntityClass) {
	sort.Sort(entitySorter(entity))
}

type entitySorter []tg.MessageEntityClass

func (e entitySorter) Len() int {
	return len(e)
}

func (e entitySorter) Less(i, j int) bool {
	a, b := e[i], e[j]
	return a.GetOffset() < b.GetOffset() ||
		a.GetLength() > b.GetLength()
}

func (e entitySorter) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

