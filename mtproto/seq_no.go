package mtproto

// seqNo returns current sequence number
func (c *Conn) seqNo(content bool) int32 {
	c.sentContentMessagesMux.Lock()
	defer c.sentContentMessagesMux.Unlock()
	
	current := c.sentContentMessages * 2
	if content {
		current++
		c.sentContentMessages++
	}

	return current
}
