package socks

import "github.com/5vnetwork/vx-core/i"

func (c *Server) GetUser(username, password string) i.User {
	if len(c.users) == 0 {
		return nil
	}
	pass, found := c.users[username]
	if !found {
		return nil
	}
	if pass.Secret() == password {
		return pass
	}
	return nil
}
