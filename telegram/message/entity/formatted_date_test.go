package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFormattedDateFormat(t *testing.T) {
	const date = 1640000000

	for _, tt := range []struct {
		format  string
		want    *tg.MessageEntityFormattedDate
		wantErr bool
	}{
		{format: "", want: &tg.MessageEntityFormattedDate{Offset: 0, Length: 4, Date: date}},
		{format: "r", want: &tg.MessageEntityFormattedDate{Offset: 0, Length: 4, Relative: true, Date: date}},
		{format: "R", want: &tg.MessageEntityFormattedDate{Offset: 0, Length: 4, Relative: true, Date: date}},
		{format: "tT", want: &tg.MessageEntityFormattedDate{
			Offset: 0, Length: 4, ShortTime: true, LongTime: true, Date: date,
		}},
		{format: "dDw", want: &tg.MessageEntityFormattedDate{
			Offset: 0, Length: 4, ShortDate: true, LongDate: true, DayOfWeek: true, Date: date,
		}},
		// "r"/"R" are only valid as the whole string.
		{format: "rt", wantErr: true},
		{format: "x", wantErr: true},
	} {
		t.Run(tt.format, func(t *testing.T) {
			f, err := FormattedDateFormat(tt.format, date)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, f(0, 4))
		})
	}
}
