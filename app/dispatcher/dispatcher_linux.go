//go:build linux && !android && server

package dispatcher

// func (d *Dispatcher) recordlinkStats(ctx context.Context, info *session.Info) {
// 	linkMetrics, err := net.GetTCPConnectionRTT(info.RawConn)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed to get link metrics")
// 		return
// 	}
// 	network := net.PrefixStringFromIP(info.Source.Address.IP())
// 	stats, ok := d.LinkStats.Load(network)
// 	if !ok {
// 		stats = &LinkStats{}
// 		d.LinkStats.Store(network, stats)
// 	}
// 	ls := stats.(*LinkStats)

// 	ls.Lock()
// 	ls.Num++
// 	ls.BWTotal += linkMetrics.Bandwidth
// 	ls.PingTotal += linkMetrics.Rtt
// 	ls.Unlock()

// 	log.Debug().Str("src prefix", network).
// 		Uint32("bw", linkMetrics.Bandwidth).
// 		Uint32("rtt", linkMetrics.Rtt).Msg("link stats")
// }
