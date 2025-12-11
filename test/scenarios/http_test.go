//go:build test

package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/serial"
	httptest "github.com/5vnetwork/vx-core/test/servers/http"
	"github.com/5vnetwork/vx-core/test/servers/tcp"

	"github.com/google/go-cmp/cmp"
)

func TestHttpConformance(t *testing.T) {
	httpServerPort := tcp.PickPort()
	httpServer := &httptest.Server{
		Port:        httpServerPort,
		PathHandler: make(map[string]http.HandlerFunc),
	}
	_, err := httpServer.Start()
	common.Must(err)
	defer httpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Log: &configs.LoggerConfig{
			LogLevel: configs.Level_DEBUG,
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port:     uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxyconfig.HttpServerConfig{}),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://127.0.0.1:" + httpServerPort.String())
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("status: ", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		common.Must(err)
		if string(content) != "Home" {
			t.Fatal("body: ", string(content))
		}
	}
}

func TestHttpError(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: func(msg []byte) []byte {
			return []byte{}
		},
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	time.AfterFunc(Timeout, func() {
		tcpServer.ShouldClose = true
	})

	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port:     uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxyconfig.HttpServerConfig{}),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://127.0.0.1:" + dest.Port.String())
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 503 {
			t.Error("status: ", resp.StatusCode)
		}
	}
}

func TestHttpPost(t *testing.T) {
	httpServerPort := tcp.PickPort()
	httpServer := &httptest.Server{
		Port: httpServerPort,
		PathHandler: map[string]http.HandlerFunc{
			"/testpost": func(w http.ResponseWriter, r *http.Request) {
				payload, err := buf.ReadAllToBytes(r.Body)
				r.Body.Close()
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte("Unable to read all payload"))
					return
				}
				payload = Xor(payload)
				w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
				w.Write(payload)
			},
		},
	}

	_, err := httpServer.Start()
	common.Must(err)
	defer httpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port:     uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxyconfig.HttpServerConfig{}),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		payload := make([]byte, 102400)
		common.Must2(rand.Read(payload))

		resp, err := client.Post("http://127.0.0.1:"+httpServerPort.String()+"/testpost", "application/x-www-form-urlencoded", bytes.NewReader(payload))
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("status: ", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		common.Must(err)
		if r := cmp.Diff(content, Xor(payload)); r != "" {
			t.Fatal(r)
		}
	}
}

func setProxyBasicAuth(req *http.Request, user, pass string) {
	req.SetBasicAuth(user, pass)
	req.Header.Set("Proxy-Authorization", req.Header.Get("Authorization"))
	req.Header.Del("Authorization")
}

// func TestHttpBasicAuth(t *testing.T) {
// 	httpServerPort := tcp.PickPort()
// 	httpServer := &httptest.Server{
// 		Port:        httpServerPort,
// 		PathHandler: make(map[string]http.HandlerFunc),
// 	}
// 	_, err := httpServer.Start()
// 	common.Must(err)
// 	defer httpServer.Close()

// 	serverPort := tcp.PickPort()
// 	serverConfig := &configs.TmConfig{
// 		InboundManager: &configs.InboundManagerConfig{
// 			Handlers: []*anypb.Any{
// 				serial.ToTypedMessage(&configs.ProxyInboundConfig{
// 					Port:     uint32(serverPort),
// 					Protocol: serial.ToTypedMessage(&proxyconfig.HttpServerConfig{}),
// 				}),
// 			},
// 		},
// 	}

// 	// Accounts: map[string]string{
// 	// 	"a": "b",
// 	// },

// 	server, err := x.NewInstanceTM(serverConfig)
// 	common.Must(err)

// 	common.Must(server.Start())
// 	defer server.Close()

// 	{
// 		transport := &http.Transport{
// 			Proxy: func(req *http.Request) (*url.URL, error) {
// 				return url.Parse("http://127.0.0.1:" + serverPort.String())
// 			},
// 		}

// 		client := &http.Client{
// 			Transport: transport,
// 		}

// 		{
// 			resp, err := client.Get("http://127.0.0.1:" + httpServerPort.String())
// 			common.Must(err)
// 			defer resp.Body.Close()
// 			if resp.StatusCode != 407 {
// 				t.Fatal("status: ", resp.StatusCode)
// 			}
// 		}

// 		{
// 			ctx := context.Background()
// 			req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:"+httpServerPort.String(), nil)
// 			common.Must(err)

// 			setProxyBasicAuth(req, "a", "c")
// 			resp, err := client.Do(req)
// 			common.Must(err)
// 			defer resp.Body.Close()
// 			if resp.StatusCode != 407 {
// 				t.Fatal("status: ", resp.StatusCode)
// 			}
// 		}

// 		{
// 			ctx := context.Background()
// 			req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:"+httpServerPort.String(), nil)
// 			common.Must(err)

// 			setProxyBasicAuth(req, "a", "b")
// 			resp, err := client.Do(req)
// 			common.Must(err)
// 			defer resp.Body.Close()
// 			if resp.StatusCode != 200 {
// 				t.Fatal("status: ", resp.StatusCode)
// 			}

// 			content, err := io.ReadAll(resp.Body)
// 			common.Must(err)
// 			if string(content) != "Home" {
// 				t.Fatal("body: ", string(content))
// 			}
// 		}
// 	}
// }
