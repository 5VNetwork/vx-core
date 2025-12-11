//go:build !confonly
// +build !confonly

package vmess

import (
	"github.com/5vnetwork/vx-core/common/dice"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/uuid"
)

// MemoryAccount is an in-memory form of VMess account.
type MemoryAccount struct {
	UserId string
	// ID is the main ID of the account.
	ID *protocol.ID
	// AlterIDs are the alternative IDs of the account.
	AlterIDs []*protocol.ID
	// Security type of the account. Used for client connections.
	Security                      protocol.SecurityType
	AuthenticatedLengthExperiment bool
	NoTerminationSignal           bool
}

func NewMemoryAccount(uid string, secret string, alterId uint16, security protocol.SecurityType,
	authLenExp, noTermi bool) *MemoryAccount {
	protoID := protocol.NewID(uuid.StringToUUID(secret))
	var AuthenticatedLength, NoTerminationSignal bool
	return &MemoryAccount{
		UserId:                        uid,
		Security:                      security.GetSecurityType(),
		AlterIDs:                      protocol.NewAlterIDs(protoID, alterId),
		ID:                            protoID,
		AuthenticatedLengthExperiment: AuthenticatedLength,
		NoTerminationSignal:           NoTerminationSignal,
	}
}

// AnyValidID returns an ID that is either the main ID or one of the alternative IDs if any.
func (a *MemoryAccount) AnyValidID() *protocol.ID {
	if len(a.AlterIDs) == 0 {
		return a.ID
	}
	return a.AlterIDs[dice.Roll(len(a.AlterIDs))]
}

// Equals implements protocol.Account.
func (a *MemoryAccount) Equals(account protocol.Account) bool {
	vmessAccount, ok := account.(*MemoryAccount)
	if !ok {
		return false
	}
	// TODO: handle AlterIds difference
	return a.ID.Equals(vmessAccount.ID)
}
