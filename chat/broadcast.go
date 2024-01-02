package chat

func (c *Chat) broadcast(message Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	actives := make([]*User, 0, 10)
	c.activeCount = 0
	for _, user := range c.active {
		if user.Dead {
			// tirar esse daqui....
			close(user.Channel)
			continue
		}
		c.activeCount++
		user.Channel <- message
		actives = append(actives, user)
	}
	c.active = actives
}
