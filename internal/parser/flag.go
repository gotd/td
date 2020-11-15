package parser

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

// Flag describes conditional parameter.
type Flag struct {
	// Name of the parameter.
	Name string `json:"name"`
	// Index represent bit index.
	Index int `json:"index"`
}

func (f Flag) String() string {
	return fmt.Sprintf("%s.%d", f.Name, f.Index)
}

func (f *Flag) Parse(s string) error {
	pos := strings.Index(s, ".")
	if pos < 1 {
		return xerrors.New("bad flag")
	}
	f.Name = s[:pos]
	if !isValidName(f.Name) {
		return xerrors.Errorf("name %q is invalid", f.Name)
	}
	idx, err := strconv.Atoi(s[pos+1:])
	if err != nil {
		return xerrors.New("bad index")
	}
	f.Index = idx
	return nil
}
