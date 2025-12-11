//go:build !linux || android

package dispatcher

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

func (d *Dispatcher) recordlinkStats(ctx context.Context, info *session.Info) {
}
