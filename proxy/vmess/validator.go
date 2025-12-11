package vmess

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash/crc64"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/dice"
	"github.com/5vnetwork/vx-core/common/protocol"

	"github.com/5vnetwork/vx-core/common/serial"

	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/proxy/vmess/aead"
)

const (
	updateInterval   = 10 * time.Second
	cacheDurationSec = 120
)

type userT2 struct {
	user    *MemoryAccount
	lastSec protocol.Timestamp
}

// TimedUserValidator is a user Validator based on time.
type TimedUserValidator struct {
	sync.RWMutex
	users    []*userT2
	userHash map[[16]byte]indexTimePair
	hasher   protocol.IDHash
	baseTime protocol.Timestamp
	task     *task.Periodic

	behaviorSeed  uint64
	behaviorFused bool

	aeadDecoderHolder *aead.AuthIDDecoderHolder

	legacyWarnShown bool
}

type indexTimePair struct {
	user    *userT2
	timeInc uint32

	taintedFuse *uint32
}

// NewTimedUserValidator creates a new TimedUserValidator.
func NewTimedUserValidator(hasher protocol.IDHash) *TimedUserValidator {
	tuv := &TimedUserValidator{
		users:             make([]*userT2, 0, 16),
		userHash:          make(map[[16]byte]indexTimePair, 1024),
		hasher:            hasher,
		baseTime:          protocol.Timestamp(time.Now().Unix() - cacheDurationSec*2),
		aeadDecoderHolder: aead.NewAuthIDDecoderHolder(),
	}
	tuv.task = &task.Periodic{
		Interval: updateInterval,
		Execute: func() error {
			tuv.updateUserHash()
			return nil
		},
	}
	common.Must(tuv.task.Start())
	return tuv
}

// visible for testing
func (v *TimedUserValidator) GetBaseTime() protocol.Timestamp {
	return v.baseTime
}

func (v *TimedUserValidator) generateNewHashes(nowSec protocol.Timestamp, user *userT2) {
	var hashValue [16]byte
	genEndSec := nowSec + cacheDurationSec
	genHashForID := func(id *protocol.ID) {
		idHash := v.hasher(id.Bytes())
		genBeginSec := user.lastSec
		if genBeginSec < nowSec-cacheDurationSec {
			genBeginSec = nowSec - cacheDurationSec
		}
		for ts := genBeginSec; ts <= genEndSec; ts++ {
			common.Must2(serial.WriteUint64(idHash, uint64(ts)))
			idHash.Sum(hashValue[:0])
			idHash.Reset()

			v.userHash[hashValue] = indexTimePair{
				user:        user,
				timeInc:     uint32(ts - v.baseTime),
				taintedFuse: new(uint32),
			}
		}
	}

	account := user.user

	genHashForID(account.ID)
	for _, id := range account.AlterIDs {
		genHashForID(id)
	}
	user.lastSec = genEndSec
}

func (v *TimedUserValidator) removeExpiredHashes(expire uint32) {
	for key, pair := range v.userHash {
		if pair.timeInc < expire {
			delete(v.userHash, key)
		}
	}
}

func (v *TimedUserValidator) updateUserHash() {
	now := time.Now()
	nowSec := protocol.Timestamp(now.Unix())

	v.Lock()
	defer v.Unlock()

	for _, user := range v.users {
		v.generateNewHashes(nowSec, user)
	}

	expire := protocol.Timestamp(now.Unix() - cacheDurationSec)
	if expire > v.baseTime {
		v.removeExpiredHashes(uint32(expire - v.baseTime))
	}
}

func (v *TimedUserValidator) Add(u *MemoryAccount) error {
	v.Lock()
	defer v.Unlock()

	nowSec := time.Now().Unix()

	uu := &userT2{
		user:    u,
		lastSec: protocol.Timestamp(nowSec - cacheDurationSec),
	}
	v.users = append(v.users, uu)
	v.generateNewHashes(protocol.Timestamp(nowSec), uu)

	account := uu.user
	if !v.behaviorFused {
		hashkdf := hmac.New(sha256.New, []byte("VMESSBSKDF"))
		hashkdf.Write(account.ID.Bytes())
		v.behaviorSeed = crc64.Update(v.behaviorSeed, crc64.MakeTable(crc64.ECMA), hashkdf.Sum(nil))
	}

	var cmdkeyfl [16]byte
	copy(cmdkeyfl[:], account.ID.CmdKey())
	v.aeadDecoderHolder.AddUser(cmdkeyfl, u)

	return nil
}

func (v *TimedUserValidator) Get(userHash []byte) (*MemoryAccount, protocol.Timestamp, bool, error) {
	v.RLock()
	defer v.RUnlock()

	v.behaviorFused = true

	var fixedSizeHash [16]byte
	copy(fixedSizeHash[:], userHash)
	pair, found := v.userHash[fixedSizeHash]
	if found {
		user := pair.user.user
		if atomic.LoadUint32(pair.taintedFuse) == 0 {
			return user, protocol.Timestamp(pair.timeInc) + v.baseTime, true, nil
		}
		return nil, 0, false, ErrTainted
	}
	return nil, 0, false, ErrNotFound
}

func (v *TimedUserValidator) GetAEAD(userHash []byte) (*MemoryAccount, bool, error) {
	v.RLock()
	defer v.RUnlock()

	var userHashFL [16]byte
	copy(userHashFL[:], userHash)

	userd, err := v.aeadDecoderHolder.Match(userHashFL)
	if err != nil {
		return nil, false, err
	}
	return userd.(*MemoryAccount), true, err
}

func (v *TimedUserValidator) Remove(secret string) bool {
	v.Lock()
	defer v.Unlock()

	secret = strings.ToLower(secret)
	idx := -1
	for i, u := range v.users {
		if strings.EqualFold(u.user.ID.String(), secret) {
			idx = i
			var cmdkeyfl [16]byte
			copy(cmdkeyfl[:], u.user.ID.CmdKey())
			v.aeadDecoderHolder.RemoveUser(cmdkeyfl)
			break
		}
	}
	if idx == -1 {
		return false
	}
	ulen := len(v.users)

	v.users[idx] = v.users[ulen-1]
	v.users[ulen-1] = nil
	v.users = v.users[:ulen-1]

	return true
}

func (v *TimedUserValidator) RemoveByUid(uid string) bool {
	v.Lock()
	defer v.Unlock()

	idx := -1
	for i, u := range v.users {
		if u.user.UserId == uid {
			idx = i
			var cmdkeyfl [16]byte
			copy(cmdkeyfl[:], u.user.ID.CmdKey())
			v.aeadDecoderHolder.RemoveUser(cmdkeyfl)
			break
		}
	}
	if idx == -1 {
		return false
	}
	ulen := len(v.users)

	v.users[idx] = v.users[ulen-1]
	v.users[ulen-1] = nil
	v.users = v.users[:ulen-1]

	return true
}

// Close implements common.Closable.
func (v *TimedUserValidator) Close() error {
	return v.task.Close()
}

func (v *TimedUserValidator) GetBehaviorSeed() uint64 {
	v.Lock()
	defer v.Unlock()

	v.behaviorFused = true
	if v.behaviorSeed == 0 {
		v.behaviorSeed = dice.RollUint64()
	}
	return v.behaviorSeed
}

func (v *TimedUserValidator) BurnTaintFuse(userHash []byte) error {
	v.RLock()
	defer v.RUnlock()

	var userHashFL [16]byte
	copy(userHashFL[:], userHash)

	pair, found := v.userHash[userHashFL]
	if found {
		if atomic.CompareAndSwapUint32(pair.taintedFuse, 0, 1) {
			return nil
		}
		return ErrTainted
	}
	return ErrNotFound
}

/*
	ShouldShowLegacyWarn will return whether a Legacy Warning should be shown

Not guaranteed to only return true once for every inbound, but it is okay.
*/
func (v *TimedUserValidator) ShouldShowLegacyWarn() bool {
	if v.legacyWarnShown {
		return false
	}
	v.legacyWarnShown = true
	return true
}

var ErrNotFound = errors.New("Not Found")

var ErrTainted = errors.New("ErrTainted")
