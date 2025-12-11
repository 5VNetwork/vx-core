package trojan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/5vnetwork/vx-core/common"
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct {
	Uid      string
	Password []byte
	Key      []byte
}

func NewMemoryAccount(u string, password string) *MemoryAccount {
	return &MemoryAccount{
		Uid:      u,
		Password: []byte(password),
		Key:      hexSha224([]byte(password)),
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
