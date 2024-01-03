package chat

func (c *Chat) broadcast(message Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.history = append(c.history, &message)
	if len(c.history) == c.HistorySize+1 {
		c.history = c.history[1:c.HistorySize]
	}

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
