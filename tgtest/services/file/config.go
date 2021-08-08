package file

type Config struct {
	// Storage to store files.
	// InMemory will be used.
	Storage Storage
	// HashPartSize is a size of part to use in tg.FileHash.
	HashPartSize int
	// HashRangeSize is size of range to return in upload.getFileHashes.
	HashRangeSize int
}

func (c *Config) setDefaults() {
	if c.Storage == nil {
		c.Storage = NewInMemory()
	}
	// Telegram usually uses this values.
	if c.HashPartSize == 0 {
		c.HashPartSize = 131072
	}
	if c.HashRangeSize == 0 {
		c.HashRangeSize = 10
	}
}
