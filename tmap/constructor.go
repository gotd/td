package tmap

import "github.com/gotd/td/bin"

// Constructor maps type id to type constructor.
type Constructor struct {
	types map[uint32]func() bin.Object
}

func (c *Constructor) add(mapping map[uint32]func() bin.Object) {
	// Probably it is good place to catch collisions, but
	// ignoring for now.
	for k, v := range mapping {
		c.types[k] = v
	}
}

// New instantiates new value for type id or returns nil.
func (c *Constructor) New(id uint32) bin.Object {
	if c == nil || len(c.types) == 0 {
		return nil
	}
	fn := c.types[id]
	if fn == nil {
		return nil
	}
	return fn()
}

// NewConstructor merges mappings into Constructor.
func NewConstructor(mappings ...map[uint32]func() bin.Object) *Constructor {
	c := &Constructor{
		types: map[uint32]func() bin.Object{},
	}
	for _, m := range mappings {
		c.add(m)
	}
	return c
}
