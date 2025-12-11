package clientgrpc

// TODO
// func (s *ClientGrpc) GetOutboundStatsStream(req *GetOutboundStatsRequest,
// 	stream grpc.ServerStreamingServer[OutboundStats]) error {
// 	ticker := time.NewTicker(time.Second * time.Duration(req.Interval))
// 	defer ticker.Stop()
// 	// st := s.getStats()
// 	for {
// 		select {
// 		case <-stream.Context().Done():
// 			return nil
// 		case <-ticker.C:
// 			// st.OutboundStats.Range(func(key, value any) bool {
// 			// 	tag := key.(string)
// 			// 	o := &OutboundStats{
// 			// 		Up:   os.UpCounter.Swap(0),
// 			// 		Down: os.DownCounter.Swap(0),
// 			// 		Rate: os.Throughput.Load(),
// 			// 		Ping: os.Ping.Load(),
// 			// 		Id:   tag,
// 			// 	}
// 			// 	if err := stream.Send(o); err != nil {
// 			// 		return false
// 			// 	}
// 			// 	return true
// 			// })
// 		}
// 	}
// }
