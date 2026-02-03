// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package user

import (
	"sync"
	"sync/atomic"
)

type Manager struct {
	sync.RWMutex
	Users map[string]*User
}

type User struct {
	uuid    string
	secret  string
	counter atomic.Uint64
}

func NewUser(uid string, secret string) *User {
	return &User{
		uuid:   uid,
		secret: secret,
	}
}

func (u *User) Uid() string {
	return u.uuid
}

func (u *User) Secret() string {
	return u.secret
}

func (u *User) Counter() *atomic.Uint64 {
	return &u.counter
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
