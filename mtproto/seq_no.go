package mtproto

// seqNo returns current sequence number
func (c *Client) seqNo() int32 {
	c.sentContentMessagesMux.Lock()
	current := c.sentContentMessages * 2
	c.sentContentMessages++
	c.sentContentMessagesMux.Unlock()

	return current
}
