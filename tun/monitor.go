package tun

import (
	sync "sync"

	"slices"

	"github.com/5vnetwork/vx-core/i"
)

type DefaultInterfaceChangeNotifier struct {
	lock      sync.RWMutex
	observers []i.DefaultInterfaceChangeObserver
}

func (n *DefaultInterfaceChangeNotifier) Register(observer i.DefaultInterfaceChangeObserver) {
	n.lock.Lock()
	n.observers = append(n.observers, observer)
	n.lock.Unlock()
}

func (n *DefaultInterfaceChangeNotifier) Unregister(observer i.DefaultInterfaceChangeObserver) {
	n.lock.Lock()
	defer n.lock.Unlock()
	for i, o := range n.observers {
		if o == observer {
			n.observers = slices.Delete(n.observers, i, i+1)
			break
		}
	}
}

func (n *DefaultInterfaceChangeNotifier) Notify() {
	n.lock.RLock()
	defer n.lock.RUnlock()
	for _, o := range n.observers {
		go o.OnDefaultInterfaceChanged()
	}
}
