package dispatcher

// import (
// 	"sync"
// 	"github.com/5vnetwork/vx-core/common/buf"
// 	"github.com/5vnetwork/vx-core/i"
// )

// type FlowPacketDispatcher struct {
// 	sync.RWMutex
// 	handler            i.FlowHandler
// 	flowHandlerRunning bool
// }

// func

// func (f *FlowPacketDispatcher) WriteMultiBuffer(mb buf.MultiBuffer) error {
// 	f.Lock()
// 	defer f.Unlock()
// 	if !f.flowHandlerRunning {
// 		go func() {
// 			f.flowHandlerRunning = true
// 			err := f.handler.HandleFlow(mb)
// 		}()
// 	}

// }

// func (f *FlowPacketDispatcher) ReadMultiBuffer() (buf.MultiBuffer, error) {
// 	return nil, nil
// }

// func (f *FlowPacketDispatcher) Start() error {
// 	return nil
// }

// func (f *FlowPacketDispatcher) Close() error {
// 	return nil
// }
