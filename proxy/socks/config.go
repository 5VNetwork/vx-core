package socks

func (c *Server) HasAccount(username, password string) bool {
	if len(c.users) == 0 {
		return false
	}
	pass, found := c.users[username]
	if !found {
		return false
	}
	return pass == password
}
