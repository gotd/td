package entity

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestSortEntities(t *testing.T) {
	for _, tt := range []struct {
		name   string
		input  []tg.MessageEntityClass
		expect []tg.MessageEntityClass
	}{
		{
			name: "OffsetAscending",
			input: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 5, Length: 1},
				&tg.MessageEntityItalic{Offset: 0, Length: 1},
				&tg.MessageEntityCode{Offset: 3, Length: 10},
			},
			expect: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 0, Length: 1},
				&tg.MessageEntityCode{Offset: 3, Length: 10},
				&tg.MessageEntityBold{Offset: 5, Length: 1},
			},
		},
		{
			name: "LengthDescending",
			input: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 0, Length: 1},
				&tg.MessageEntityItalic{Offset: 0, Length: 5},
			},
			expect: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 0, Length: 5},
				&tg.MessageEntityBold{Offset: 0, Length: 1},
			},
		},
		{
			name: "TypePriority",
			input: []tg.MessageEntityClass{
				&tg.MessageEntityCustomEmoji{Offset: 0, Length: 2},
				&tg.MessageEntitySpoiler{Offset: 0, Length: 2},
				&tg.MessageEntityItalic{Offset: 0, Length: 2},
				&tg.MessageEntityBold{Offset: 0, Length: 2},
				&tg.MessageEntityMention{Offset: 0, Length: 2},
				&tg.MessageEntityTextURL{Offset: 0, Length: 2},
				&tg.MessageEntityCode{Offset: 0, Length: 2},
				&tg.MessageEntityPre{Offset: 0, Length: 2},
				&tg.MessageEntityPre{Offset: 0, Length: 2, Language: "go"},
				&tg.MessageEntityBlockquote{Offset: 0, Length: 2},
			},
			expect: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 0, Length: 2},
				&tg.MessageEntityPre{Offset: 0, Length: 2, Language: "go"},
				&tg.MessageEntityPre{Offset: 0, Length: 2},
				&tg.MessageEntityCode{Offset: 0, Length: 2},
				&tg.MessageEntityTextURL{Offset: 0, Length: 2},
				&tg.MessageEntityMention{Offset: 0, Length: 2},
				&tg.MessageEntityBold{Offset: 0, Length: 2},
				&tg.MessageEntityItalic{Offset: 0, Length: 2},
				&tg.MessageEntitySpoiler{Offset: 0, Length: 2},
				&tg.MessageEntityCustomEmoji{Offset: 0, Length: 2},
			},
		},
		{
			name: "Stable",
			input: []tg.MessageEntityClass{
				&tg.MessageEntityURL{Offset: 0, Length: 2},
				&tg.MessageEntityHashtag{Offset: 0, Length: 2},
				&tg.MessageEntityMention{Offset: 0, Length: 2},
			},
			expect: []tg.MessageEntityClass{
				&tg.MessageEntityURL{Offset: 0, Length: 2},
				&tg.MessageEntityHashtag{Offset: 0, Length: 2},
				&tg.MessageEntityMention{Offset: 0, Length: 2},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			SortEntities(tt.input)
			require.Equal(t, tt.expect, tt.input)
		})
	}
}

func TestSortEntitiesProperty(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for iter := 0; iter < 1000; iter++ {
		n := 2 + r.Intn(20)
		es := make([]tg.MessageEntityClass, 0, n)
		for i := 0; i < n; i++ {
			es = append(es, &tg.MessageEntityBold{Offset: r.Intn(50), Length: 1 + r.Intn(20)})
		}
		SortEntities(es)
		for i := 1; i < len(es); i++ {
			prev, cur := es[i-1], es[i]
			require.LessOrEqual(t, prev.GetOffset(), cur.GetOffset(), "offsets: %v", es)
			if prev.GetOffset() == cur.GetOffset() {
				require.GreaterOrEqual(t, prev.GetLength(), cur.GetLength(), "lengths: %v", es)
			}
		}
	}
}

func TestSortEntitiesAllocs(t *testing.T) {
	es := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 5, Length: 1},
		&tg.MessageEntityItalic{Offset: 0, Length: 1},
		&tg.MessageEntityCode{Offset: 3, Length: 10},
	}
	allocs := testing.AllocsPerRun(100, func() {
		es[0], es[2] = es[2], es[0]
		SortEntities(es)
	})
	require.Zero(t, allocs)
}
