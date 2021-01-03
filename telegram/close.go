package telegram

// Close closes underlying connection.
func (c *Client) Close() error {
	c.cancel()

	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}
