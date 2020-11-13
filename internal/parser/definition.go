package parser

import (
	"hash/crc32"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

type Field struct {
	Name string
	Type string
}

// Definition represents TL type definition.
type Definition struct {
	Name      string
	ID        uint32 // crc32(definition) or explicitly specified
	Fields    []Field
	Interface string
}

func parseDefinition(line string) (Definition, error) {
	line = strings.TrimSuffix(line, ";")
	d := Definition{}
	// peerUser#9db1bc6d user_id:int = Peer;
	// name#ID flags = Interface;
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return Definition{}, xerrors.New("unexpected line elems")
	}
	{
		// Parsing interface name.
		last := parts[len(parts)-1]
		d.Interface = last
	}
	{
		// Parsing definition name and id.
		first := parts[0]
		nameParts := strings.SplitN(first, tokID, 2)
		d.Name = nameParts[0]
		if len(nameParts) > 1 {
			idHex := nameParts[1]
			id, err := strconv.ParseInt(idHex, 16, 32)
			if err != nil {
				return Definition{}, xerrors.Errorf("%s is invalid id: %w", idHex, id)
			}
			d.ID = uint32(id)
		} else {
			// Automatically computing.
			d.ID = crc32.ChecksumIEEE([]byte(line))
		}
	}
	for i, f := range parts[1 : len(parts)-2] {
		// Parsing fields.
		fieldParts := strings.SplitN(f, ":", 2)
		if len(fieldParts) != 2 {
			return d, xerrors.Errorf("field %i: unexpected parts count", i)
		}
		d.Fields = append(d.Fields, Field{
			Name: fieldParts[0],
			Type: fieldParts[1],
		})
	}
	return d, nil
}
