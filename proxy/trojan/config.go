package trojan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct {
	User     i.User
	Password []byte
	Key      []byte
}

func NewMemoryAccount(user i.User) *MemoryAccount {
	return &MemoryAccount{
		User:     user,
		Password: []byte(user.Secret()),
		Key:      hexSha224([]byte(user.Secret())),
	}
}

func hexSha224(password []byte) []byte {
	buf := make([]byte, 56)
	hash := sha256.New224()
	common.Must2(hash.Write(password))
	hex.Encode(buf, hash.Sum(nil))
	return buf
}

func hexString(data []byte) string {
	str := ""
	for _, v := range data {
		str += fmt.Sprintf("%02x", v)
	}
	return str
}
