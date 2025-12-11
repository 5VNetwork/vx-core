package user

import (
	"sync"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/set"
)

type Manager struct {
	sync.RWMutex
	Users map[string]*User
}

type User struct {
	uuid     string
	level    uint32
	secret   string
	counter  atomic.Uint64
	prefexes set.Set[string]
}

func NewUser(uid string, level uint32, secret string) *User {
	return &User{
		uuid:     uid,
		level:    level,
		secret:   secret,
		prefexes: set.NewSet[string](),
	}
}

func (u *User) Uid() string {
	return u.uuid
}

func (u *User) Level() uint32 {
	return u.level
}

func (u *User) Secret() string {
	return u.secret
}

func (u *User) Counter() *atomic.Uint64 {
	return &u.counter
}

func (u *User) GetPrefixesNum() int {
	slice := u.prefexes.ToSlice()
	u.prefexes.Clear()
	return len(slice)
}

func (u *User) AddPrefix(prefix string) {
	u.prefexes.Add(prefix)
}

func NewManager() *Manager {
	return &Manager{
		Users: make(map[string]*User),
	}
}

func (m *Manager) AddUser(u *User) {
	m.Lock()
	defer m.Unlock()
	if exsiting, ok := m.Users[u.uuid]; ok {
		exsiting.level = u.level
		exsiting.secret = u.secret
		return
	}
	m.Users[u.uuid] = u
}

func (m *Manager) Number() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.Users)
}

func (m *Manager) AllUsers() []*User {
	m.RLock()
	defer m.RUnlock()
	users := make([]*User, 0, len(m.Users))
	for _, u := range m.Users {
		users = append(users, u)
	}
	return users
}

func (m *Manager) GetUser(id string) *User {
	m.RLock()
	defer m.RUnlock()
	return m.Users[id]
}

func (m *Manager) RemoveUser(id string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Users, id)
}
