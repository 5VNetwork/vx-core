package vless

import (
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
)

// MemoryAccount is an in-memory form of VLess account.
type MemoryAccount struct {
	User i.User
	// ID of the account.
	ID *protocol.ID
	// Flow of the account. May be "xtls-rprx-vision".
	Flow string
	// Encryption of the account. Used for client connections, and only accepts "none" for now.
	Encryption string
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(account protocol.Account) bool {
	vlessAccount, ok := account.(*MemoryAccount)
	if !ok {
		return false
	}
	return a.ID.Equals(vlessAccount.ID)
}
