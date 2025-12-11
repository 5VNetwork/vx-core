package api

// func ApiHandlerIp(req *HandlerIpRequest, dialerFactory transport.DialerFactory) (*HandlerIpResponse, error) {
// 	h, err := create.NewOutHandler(req.Handler,
// 		dialerFactory, policy.New(), stats.NewStats(), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ret := &HandlerIpResponse{}

// 	// ipv4
// 	httpClient := outbound.HandlerToHttpClient(h)
// 	httpClient.Timeout = 5 * time.Second
// 	rsp, err := httpClient.Get("https://ip4only.me/api/")
// 	if err == nil {
// 		data, err := io.ReadAll(rsp.Body)
// 		rsp.Body.Close()
// 		if err == nil {
// 			log.Println(string(data))
// 			splitted := strings.Split(string(data), ",")
// 			ret.Ip4 = splitted[1]
// 		}
// 	}

// 	// ipv6
// 	httpClient = outbound.HandlerToHttpClient(h)
// 	httpClient.Timeout = 5 * time.Second
// 	rsp, err = httpClient.Get("https://ip6only.me/api/")
// 	if err == nil {
// 		data, err := io.ReadAll(rsp.Body)
// 		rsp.Body.Close()
// 		if err == nil {
// 			log.Println(string(data))
// 			splitted := strings.Split(string(data), ",")
// 			ret.Ip6 = splitted[1]
// 		}
// 	} else {
// 		// log.Println(err)
// 	}

// 	return ret, nil
// }

// func TestHandlerIPv6(h i.Outbound) (bool, error) {
// 	httpClient := outbound.HandlerToHttpClient(h)
// 	httpClient.Timeout = 5 * time.Second
// 	rsp, err := httpClient.Get("https://ip6only.me/api/")
// 	if err == nil {
// 		data, err := io.ReadAll(rsp.Body)
// 		rsp.Body.Close()
// 		if err == nil {
// 			log.Println(string(data))
// 			splitted := strings.Split(string(data), ",")
// 			return splitted[1] != "", nil
// 		}
// 	}
// 	return false, nil
// }
