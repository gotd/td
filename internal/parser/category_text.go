package parser

import "golang.org/x/xerrors"

func (i *Category) UnmarshalText(text []byte) error {
	for idx := range _Category_index {
		if Category(idx).String() == string(text) {
			*i = Category(idx)
			return nil
		}
	}
	return xerrors.Errorf("unknown category %q", text)
}

func (i Category) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}
