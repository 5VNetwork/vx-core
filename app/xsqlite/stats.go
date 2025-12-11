package xsqlite

// import (
// 	"github.com/5vnetwork/vx-core/app/outbound"
// 	"github.com/5vnetwork/vx-core/common/task"
// )

// // periodically save node stats to database
// type NodeStats struct {
// 	updateHandler
// 	outStats *outbound.OutStats
// 	task     *task.PeriodicTask
// }

// func NewNodeStats(updateHandler updateHandler, outStats *outbound.OutStats) *NodeStats {

// }

// func (n *NodeStats) Start() error {
// 	return n.task.Start()
// }

// func (n *NodeStats) Close() error {
// 	return n.task.Close()
// }

// func (n *NodeStats) saveToDisk() error {
// 	n.outStats.Lock()
// 	defer n.outStats.Unlock()
// 	for tag, stats := range n.outStats.Map {
// 		err := n.updateHandler.UpdateHandler(tag, map[string]interface{}{
// 			"up":   stats.UpCounter.Load(),
// 			"down": stats.DownCounter.Load(),
// 		})
// 	}
// }

// type updateHandler interface {
// 	UpdateHandler(id int, m map[string]interface{}) error
// }
