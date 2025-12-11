package trojan

import (
	"sync"

	"github.com/5vnetwork/vx-core/common/errors"
)

// Validator stores valid trojan users.
type Validator struct {
	// Considering uuidToAccount's usage here, map + sync.Mutex/RWMutex may have better performance.
	uuidToAccount   sync.Map // key: string , value: *MemoryAccount
	secretToAccount sync.Map // key: hexString(u.Key), value: *MemoryAccount
}

// Add a trojan user, Email must be empty or unique.
func (v *Validator) Add(u *MemoryAccount) {
	v.uuidToAccount.Store(u.Uid, u)
	v.secretToAccount.Store(hexString(u.Key), u)
}

// Del a trojan user with a non-empty Email.
func (v *Validator) Del(uid string) error {
	u, _ := v.uuidToAccount.Load(uid)
	if u == nil {
		return errors.New(" not found.")
	}
	v.uuidToAccount.Delete(uid)
	v.secretToAccount.Delete(hexString(u.(*MemoryAccount).Key))
	return nil
}

// Get a trojan user with hashed key, nil if user doesn't exist.
func (v *Validator) Get(hash string) *MemoryAccount {
	u, _ := v.secretToAccount.Load(hash)
	if u != nil {
		return u.(*MemoryAccount)
	}
	return nil
}
