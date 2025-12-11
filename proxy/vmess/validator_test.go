package vmess_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/proxy/vmess"
)

func TestUserManagement(t *testing.T) {
	validator := vmess.NewTimedUserValidator(protocol.DefaultIDHash)
	uid := uuid.New()
	secret := uuid.New()
	userAccount := vmess.NewMemoryAccount(uid.String(), secret.String(), 0,
		protocol.SecurityType_AUTO, false, false)
	validator.Add(userAccount)

	removed := validator.Remove(secret.String())
	if !removed {
		t.Error("Failed to remove user")
	}
}
