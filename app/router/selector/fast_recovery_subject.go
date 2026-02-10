package selector

import (
	"slices"
	"sync"
)

type FastRecoveryChangeNotifier struct {
	lock      sync.RWMutex
	observers []FastRecoveryChangeObserver
}

type FastRecoveryChangeObserver interface {
	OnFastRecoveryChanged(isRecovery bool)
}

type OnFastRecoveryChangedFunc func()

func (f OnFastRecoveryChangedFunc) OnFastRecoveryChanged() {
	f()
}

func (n *FastRecoveryChangeNotifier) RegisterFastRevoceryObserver(observer FastRecoveryChangeObserver) {
	n.lock.Lock()
	n.observers = append(n.observers, observer)
	n.lock.Unlock()
}

func (n *FastRecoveryChangeNotifier) UnregisterFastRevoceryObserver(observer FastRecoveryChangeObserver) {
	n.lock.Lock()
	defer n.lock.Unlock()
	for i, o := range n.observers {
		if o == observer {
			n.observers = slices.Delete(n.observers, i, i+1)
			break
		}
	}
}

func (n *FastRecoveryChangeNotifier) Notify(isRecovery bool) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	for _, o := range n.observers {
		go o.OnFastRecoveryChanged(isRecovery)
	}
}
