package api_test

// func TestApiDownloadToFile(t *testing.T) {
// 	request := &api.DownloadRequest{
// 		Url:  api.UsableTestUrl5,
// 		Dest: "./geoip.dat",
// 		OutboundHandlers: []*configs.OutboundHandlerConfig{
// 			{
// 				Protocol: serial.ToTypedMessage(&proxyconfig.FreedomConfig{}),
// 			},
// 		},
// 	}
// 	response := api.ApiDownload(request)
// 	if response.GetError() != "" {
// 		t.Fatal(response.GetError())
// 	}
// 	// verify the file exists
// 	if _, err := os.Stat(request.Dest); os.IsNotExist(err) {
// 		t.Fatal("file not exists")
// 	}
// 	os.Remove(request.Dest)
// }

// func TestApiDownloadToMemory(t *testing.T) {
// 	request := &api.DownloadRequest{
// 		Url: api.UsableTestUrl5,
// 		OutboundHandlers: []*configs.OutboundHandlerConfig{
// 			{
// 				Protocol: serial.ToTypedMessage(&proxyconfig.FreedomConfig{}),
// 			},
// 		},
// 	}
// 	response := api.ApiDownload(request)
// 	if response.GetError() != "" {
// 		t.Fatal(response.GetError())
// 	}
// 	t.Log(string(response.GetData()))
// }
