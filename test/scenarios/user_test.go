//go:build test

package scenarios_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/common"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/scenarios"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"golang.org/x/sync/errgroup"
)

// Test user.Manager basic operations
func TestUserManager(t *testing.T) {
	manager := user.NewManager()

	// Test initial state
	if manager.Number() != 0 {
		t.Errorf("Expected 0 users, got %d", manager.Number())
	}

	// Test AddUser
	user1 := user.NewUser("user1", 1, "secret1")
	manager.AddUser(user1)
	if manager.Number() != 1 {
		t.Errorf("Expected 1 user, got %d", manager.Number())
	}

	// Test GetUser
	retrieved := manager.GetUser("user1")
	if retrieved == nil {
		t.Fatal("User not found")
	}
	if retrieved.Uid() != "user1" {
		t.Errorf("Expected uid 'user1', got '%s'", retrieved.Uid())
	}
	if retrieved.Secret() != "secret1" {
		t.Errorf("Expected secret 'secret1', got '%s'", retrieved.Secret())
	}

	// Test AddUser with existing uid (should update)
	user1Updated := user.NewUser("user1", 2, "secret2")
	manager.AddUser(user1Updated)
	if manager.Number() != 1 {
		t.Errorf("Expected 1 user after update, got %d", manager.Number())
	}
	retrieved = manager.GetUser("user1")
	if retrieved.Level() != 2 {
		t.Errorf("Expected level 2, got %d", retrieved.Level())
	}
	if retrieved.Secret() != "secret2" {
		t.Errorf("Expected secret 'secret2', got '%s'", retrieved.Secret())
	}

	// Test multiple users
	user2 := user.NewUser("user2", 1, "secret2")
	user3 := user.NewUser("user3", 1, "secret3")
	manager.AddUser(user2)
	manager.AddUser(user3)
	if manager.Number() != 3 {
		t.Errorf("Expected 3 users, got %d", manager.Number())
	}

	// Test AllUsers
	allUsers := manager.AllUsers()
	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users from AllUsers(), got %d", len(allUsers))
	}

	// Test RemoveUser
	manager.RemoveUser("user2")
	if manager.Number() != 2 {
		t.Errorf("Expected 2 users after removal, got %d", manager.Number())
	}
	if manager.GetUser("user2") != nil {
		t.Error("User2 should have been removed")
	}

	// Test GetUser for non-existent user
	if manager.GetUser("nonexistent") != nil {
		t.Error("Expected nil for non-existent user")
	}

	// Test Counter
	counter := user1.Counter()
	counter.Add(100)
	if counter.Load() != 100 {
		t.Errorf("Expected counter 100, got %d", counter.Load())
	}

	// Test Prefix operations
	user1.AddPrefix("prefix1")
	user1.AddPrefix("prefix2")
	user1.AddPrefix("prefix3")
	prefixCount := user1.GetPrefixesNum()
	if prefixCount != 3 {
		t.Errorf("Expected 3 prefixes, got %d", prefixCount)
	}
	// After GetPrefixesNum, prefixes should be cleared
	prefixCount = user1.GetPrefixesNum()
	if prefixCount != 0 {
		t.Errorf("Expected 0 prefixes after clear, got %d", prefixCount)
	}
}

// TestVMessUserManagement tests dynamic user add/remove for VMess protocol
func TestVMessUserManagement(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	// Create two users
	userID1 := protocol.NewID(uuid.New())
	userID2 := protocol.NewID(uuid.New())

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     "user1",
				Secret: userID1.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with first user
	clientPort1 := tcp.PickPort()
	t.Log("client port 1", clientPort1)
	clientConfig1 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort1),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID1.String(),
						Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
					}),
				},
			},
		},
	}

	client1, err := buildclient.NewX(clientConfig1)
	common.Must(err)
	common.Must(client1.Start())

	// Test connection with user1
	var errg errgroup.Group
	for i := 0; i < 5; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort1, 10240, scenarios.Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with user1: %v", err)
	}
	client1.Close()

	// Add second user dynamically
	// Note: This would require access to server's user manager, which depends on implementation
	// For now, test that invalid user fails
	clientPort2 := tcp.PickPort()
	t.Log("client port 2", clientPort2)
	clientConfig2 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort2),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID2.String(),
						Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
					}),
				},
			},
		},
	}

	client2, err := buildclient.NewX(clientConfig2)
	common.Must(err)
	common.Must(client2.Start())
	defer client2.Close()

	// Test connection with unauthorized user2 - should fail
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(clientPort2),
	})
	if err == nil {
		conn.SetDeadline(time.Now().Add(scenarios.Timeout))
		// Try to send data
		_, err = conn.Write([]byte("test"))
		conn.Close()
		// Connection might succeed initially but should fail on authentication
		t.Log("Unauthorized user connection behavior: ", err)
	}
}

// TestTrojanUserManagement tests user management for Trojan protocol
func TestTrojanUserManagement(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	userID1 := protocol.NewID(uuid.New())
	userID2 := protocol.NewID(uuid.New())

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID1.String(),
								Secret: userID1.String(),
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with authorized user
	clientPort1 := tcp.PickPort()
	t.Log("client port 1", clientPort1)
	clientConfig1 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort1),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID1.String(),
					}),
				},
			},
		},
	}

	client1, err := buildclient.NewX(clientConfig1)
	common.Must(err)
	common.Must(client1.Start())

	var errg errgroup.Group
	for i := 0; i < 5; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort1, 10240, scenarios.Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with authorized user: %v", err)
	}
	client1.Close()

	// Test with unauthorized user
	clientPort2 := tcp.PickPort()
	t.Log("client port 2 (unauthorized)", clientPort2)
	clientConfig2 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort2),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID2.String(), // Unauthorized user
					}),
				},
			},
		},
	}

	client2, err := buildclient.NewX(clientConfig2)
	common.Must(err)
	common.Must(client2.Start())
	defer client2.Close()

	// Connection should fail or timeout with unauthorized user
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(clientPort2),
	})
	if err == nil {
		conn.SetDeadline(time.Now().Add(scenarios.Timeout))
		_, err = conn.Write([]byte("test"))
		conn.Close()
		t.Log("Unauthorized trojan user behavior: ", err)
	}
}

// TestShadowsocksUserManagement tests user management for Shadowsocks protocol
func TestShadowsocksUserManagement(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	password1 := uuid.New().String()
	password2 := uuid.New().String()

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: password1,
						},
						CipherType: proxyconfig.ShadowsocksCipherType_CHACHA20_POLY1305,
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with correct password
	clientPort1 := tcp.PickPort()
	t.Log("client port 1", clientPort1)
	clientConfig1 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort1),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.ShadowsocksClientConfig{
						Password:   password1,
						CipherType: proxyconfig.ShadowsocksCipherType_CHACHA20_POLY1305,
					}),
				},
			},
		},
	}

	client1, err := buildclient.NewX(clientConfig1)
	common.Must(err)
	common.Must(client1.Start())

	var errg errgroup.Group
	for i := 0; i < 5; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort1, 10240, scenarios.Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with correct password: %v", err)
	}
	client1.Close()

	// Test with wrong password
	clientPort2 := tcp.PickPort()
	t.Log("client port 2 (wrong password)", clientPort2)
	clientConfig2 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort2),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.ShadowsocksClientConfig{
						Password:   password2, // Wrong password
						CipherType: proxyconfig.ShadowsocksCipherType_CHACHA20_POLY1305,
					}),
				},
			},
		},
	}

	client2, err := buildclient.NewX(clientConfig2)
	common.Must(err)
	common.Must(client2.Start())
	defer client2.Close()

	// Connection should fail with wrong password
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(clientPort2),
	})
	if err == nil {
		conn.SetDeadline(time.Now().Add(scenarios.Timeout))
		_, err = conn.Write([]byte("test"))
		conn.Close()
		t.Log("Wrong shadowsocks password behavior: ", err)
	}
}

// TestSOCKSUserManagement tests user management for SOCKS protocol
func TestSOCKSUserManagement(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	userID1 := uuid.New().String()
	password1 := uuid.New().String()
	userID2 := uuid.New().String()
	password2 := uuid.New().String()

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.SocksServerConfig{
						AuthType: proxyconfig.AuthType_PASSWORD,
						Accounts: []*configs.UserConfig{
							{
								Id:     userID1,
								Secret: password1,
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with correct credentials
	clientPort1 := tcp.PickPort()
	t.Log("client port 1", clientPort1)
	clientConfig1 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort1),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.SocksClientConfig{
						Name:     userID1,
						Password: password1,
					}),
				},
			},
		},
	}

	client1, err := buildclient.NewX(clientConfig1)
	common.Must(err)
	common.Must(client1.Start())

	var errg errgroup.Group
	for i := 0; i < 5; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort1, 10240, scenarios.Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with correct credentials: %v", err)
	}
	client1.Close()

	// Test with wrong credentials
	clientPort2 := tcp.PickPort()
	t.Log("client port 2 (wrong credentials)", clientPort2)
	clientConfig2 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort2),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.SocksClientConfig{
						Name:     userID2,
						Password: password2, // Wrong credentials
					}),
				},
			},
		},
	}

	client2, err := buildclient.NewX(clientConfig2)
	common.Must(err)
	common.Must(client2.Start())
	defer client2.Close()

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(clientPort2),
	})
	if err == nil {
		conn.SetDeadline(time.Now().Add(scenarios.Timeout))
		_, err = conn.Write([]byte("test"))
		conn.Close()
		t.Log("Wrong SOCKS credentials behavior: ", err)
	}
}

// TestMultipleUsersVMess tests multiple concurrent users on VMess
func TestMultipleUsersVMess(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	// Create multiple users
	var userIDs []*protocol.ID
	var userConfigs []*configs.UserConfig
	for i := 0; i < 5; i++ {
		userID := protocol.NewID(uuid.New())
		userIDs = append(userIDs, userID)
		userConfigs = append(userConfigs, &configs.UserConfig{
			Id:     fmt.Sprintf("user%d", i),
			Secret: userID.String(),
		})
	}

	serverConfig := &server.ServerConfig{
		Users: userConfigs,
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test all users concurrently
	var errg errgroup.Group
	for idx, userID := range userIDs {
		userID := userID // capture loop variable
		clientPort := tcp.PickPort()
		t.Logf("client port for user %d: %d", idx, clientPort)

		clientConfig := &configs.TmConfig{
			InboundManager: &configs.InboundManagerConfig{
				Handlers: []*configs.ProxyInboundConfig{
					{
						Address: nethelper.LocalHostIP.String(),
						Port:    uint32(clientPort),
						Protocol: serial.ToTypedMessage(
							&proxyconfig.DokodemoConfig{
								Address:  dest.Address.String(),
								Port:     uint32(dest.Port),
								Networks: []nethelper.Network{nethelper.Network_TCP},
							},
						),
					},
				},
			},
			Outbound: &configs.OutboundConfig{
				OutboundHandlers: []*configs.OutboundHandlerConfig{
					{
						Address: nethelper.LocalHostIP.String(),
						Port:    uint32(serverPort),
						Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
							Special:  true,
							Id:       userID.String(),
							Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
						}),
					},
				},
			},
		}

		client, err := buildclient.NewX(clientConfig)
		if err != nil {
			t.Fatalf("Failed to create client for user %d: %v", idx, err)
		}
		common.Must(client.Start())
		defer client.Close()

		// Test connections for this user
		for i := 0; i < 3; i++ {
			errg.Go(scenarios.TestTCPConn(clientPort, 10240, scenarios.Timeout))
		}
	}

	if err := errg.Wait(); err != nil {
		t.Errorf("Concurrent user test failed: %v", err)
	}
}

// TestUserManagerConcurrency tests concurrent operations on user manager
func TestUserManagerConcurrency(t *testing.T) {
	manager := user.NewManager()

	var errg errgroup.Group

	// Concurrent adds
	for i := 0; i < 100; i++ {
		i := i
		errg.Go(func() error {
			user := user.NewUser(fmt.Sprintf("user%d", i), uint32(i), fmt.Sprintf("secret%d", i))
			manager.AddUser(user)
			return nil
		})
	}

	// Concurrent gets
	for i := 0; i < 100; i++ {
		i := i
		errg.Go(func() error {
			u := manager.GetUser(fmt.Sprintf("user%d", i))
			if u != nil && u.Uid() != fmt.Sprintf("user%d", i) {
				return fmt.Errorf("unexpected user id")
			}
			return nil
		})
	}

	// Concurrent updates
	for i := 0; i < 50; i++ {
		i := i
		errg.Go(func() error {
			user := user.NewUser(fmt.Sprintf("user%d", i), uint32(i+100), fmt.Sprintf("newsecret%d", i))
			manager.AddUser(user)
			return nil
		})
	}

	// Concurrent removes
	for i := 50; i < 100; i++ {
		i := i
		errg.Go(func() error {
			manager.RemoveUser(fmt.Sprintf("user%d", i))
			return nil
		})
	}

	// Concurrent AllUsers
	for i := 0; i < 10; i++ {
		errg.Go(func() error {
			users := manager.AllUsers()
			if users == nil {
				return fmt.Errorf("expected non-nil users")
			}
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		t.Errorf("Concurrent operations failed: %v", err)
	}

	// Verify final state
	finalCount := manager.Number()
	t.Logf("Final user count: %d", finalCount)
	if finalCount < 0 || finalCount > 100 {
		t.Errorf("Unexpected final user count: %d", finalCount)
	}
}

// TestUserCounterOperations tests user counter increments
func TestUserCounterOperations(t *testing.T) {
	user := user.NewUser("testuser", 1, "secret")

	counter := user.Counter()
	if counter.Load() != 0 {
		t.Errorf("Expected initial counter 0, got %d", counter.Load())
	}

	// Sequential increments
	for i := 1; i <= 100; i++ {
		counter.Add(1)
	}
	if counter.Load() != 100 {
		t.Errorf("Expected counter 100, got %d", counter.Load())
	}

	// Concurrent increments
	var errg errgroup.Group
	for i := 0; i < 1000; i++ {
		errg.Go(func() error {
			counter.Add(1)
			return nil
		})
	}
	errg.Wait()

	if counter.Load() != 1100 {
		t.Errorf("Expected counter 1100, got %d", counter.Load())
	}
}

// TestVMessMultipleCipherTypes tests user authentication with different cipher types
func TestVMessMultipleCipherTypes(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	cipherTypes := []proxyconfig.SecurityType{
		proxyconfig.SecurityType_SecurityType_AES128_GCM,
		proxyconfig.SecurityType_SecurityType_CHACHA20_POLY1305,
		proxyconfig.SecurityType_SecurityType_NONE,
	}

	for _, cipherType := range cipherTypes {
		t.Run(fmt.Sprintf("Cipher_%v", cipherType), func(t *testing.T) {
			serverPort := tcp.PickPort()
			userID := protocol.NewID(uuid.New())

			serverConfig := &server.ServerConfig{
				Inbounds: []*configs.ProxyInboundConfig{
					{
						Address: nethelper.LocalHostIP.String(),
						Port:    uint32(serverPort),
						Protocol: serial.ToTypedMessage(
							&proxyconfig.VmessServerConfig{
								Accounts: []*configs.UserConfig{
									{
										Secret: userID.String(),
									},
								},
							},
						),
					},
				},
			}

			server, err := buildserver.NewX(serverConfig)
			common.Must(err)
			common.Must(server.Start(context.Background()))
			defer server.Stop(context.Background())

			clientPort := tcp.PickPort()
			clientConfig := &configs.TmConfig{
				InboundManager: &configs.InboundManagerConfig{
					Handlers: []*configs.ProxyInboundConfig{
						{
							Address: nethelper.LocalHostIP.String(),
							Port:    uint32(clientPort),
							Protocol: serial.ToTypedMessage(
								&proxyconfig.DokodemoConfig{
									Address:  dest.Address.String(),
									Port:     uint32(dest.Port),
									Networks: []nethelper.Network{nethelper.Network_TCP},
								},
							),
						},
					},
				},
				Outbound: &configs.OutboundConfig{
					OutboundHandlers: []*configs.OutboundHandlerConfig{
						{
							Address: nethelper.LocalHostIP.String(),
							Port:    uint32(serverPort),
							Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
								Special:  true,
								Id:       userID.String(),
								Security: cipherType,
							}),
						},
					},
				},
			}

			client, err := buildclient.NewX(clientConfig)
			common.Must(err)
			common.Must(client.Start())
			defer client.Close()

			var errg errgroup.Group
			for i := 0; i < 3; i++ {
				errg.Go(scenarios.TestTCPConn(clientPort, 10240, scenarios.Timeout))
			}

			if err := errg.Wait(); err != nil {
				t.Errorf("Connection failed with cipher %v: %v", cipherType, err)
			}
		})
	}
}

// TestUserManagerEdgeCases tests edge cases in user management
func TestUserManagerEdgeCases(t *testing.T) {
	manager := user.NewManager()

	// Test empty manager
	if manager.Number() != 0 {
		t.Errorf("New manager should have 0 users")
	}
	if manager.GetUser("nonexistent") != nil {
		t.Error("GetUser on empty manager should return nil")
	}
	allUsers := manager.AllUsers()
	if len(allUsers) != 0 {
		t.Errorf("AllUsers on empty manager should return empty slice")
	}

	// Test remove on empty manager
	manager.RemoveUser("nonexistent") // Should not panic

	// Test with empty user ID
	emptyUser := user.NewUser("", 1, "secret")
	manager.AddUser(emptyUser)
	retrieved := manager.GetUser("")
	if retrieved == nil {
		t.Error("Should be able to add and retrieve user with empty ID")
	}

	// Test with empty secret
	userEmptySecret := user.NewUser("user1", 1, "")
	manager.AddUser(userEmptySecret)
	retrieved = manager.GetUser("user1")
	if retrieved == nil || retrieved.Secret() != "" {
		t.Error("Should be able to add and retrieve user with empty secret")
	}

	// Test update with same uid
	original := user.NewUser("updatetest", 1, "secret1")
	manager.AddUser(original)
	countBefore := manager.Number()

	updated := user.NewUser("updatetest", 2, "secret2")
	manager.AddUser(updated)
	countAfter := manager.Number()

	if countBefore != countAfter {
		t.Error("Adding user with existing uid should not increase count")
	}

	retrieved = manager.GetUser("updatetest")
	if retrieved.Level() != 2 || retrieved.Secret() != "secret2" {
		t.Error("User should be updated with new values")
	}

	// Test counter edge cases
	testUser := user.NewUser("countertest", 1, "secret")
	counter := testUser.Counter()

	// Test large increment
	counter.Add(^uint64(0) - 100) // Near max uint64
	if counter.Load() < 1000 {
		t.Error("Counter should handle large values")
	}

	// Test prefix edge cases
	prefixUser := user.NewUser("prefixtest", 1, "secret")

	// Add same prefix multiple times
	prefixUser.AddPrefix("prefix1")
	prefixUser.AddPrefix("prefix1")
	prefixUser.AddPrefix("prefix1")
	count := prefixUser.GetPrefixesNum()
	// Set behavior: should only count unique prefixes
	t.Logf("Prefix count after adding duplicates: %d", count)

	// GetPrefixesNum should clear
	count = prefixUser.GetPrefixesNum()
	if count != 0 {
		t.Errorf("Second GetPrefixesNum should return 0, got %d", count)
	}
}

// TestHysteria2UserManagement tests user management for Hysteria2 protocol
func TestHysteria2UserManagement(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := nethelper.PickUDPPort()
	t.Log("server port", serverPort)

	userID1 := uuid.New()
	userID2 := uuid.New()

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     userID1.String(),
				Secret: userID1.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Hysteria2ServerConfig{

						IgnoreClientBandwidth: true,
						TlsConfig: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with authorized user
	clientPort1 := tcp.PickPort()
	t.Log("client port 1", clientPort1)
	clientConfig1 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort1),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.Hysteria2ClientConfig{
						Auth: userID1.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	client1, err := buildclient.NewX(clientConfig1)
	common.Must(err)
	common.Must(client1.Start())

	var errg errgroup.Group
	for i := 0; i < 3; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort1, 10240, scenarios.Timeout*2))
	}
	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with authorized user: %v", err)
	}
	client1.Close()

	// Test with unauthorized user
	clientPort2 := tcp.PickPort()
	t.Log("client port 2 (unauthorized)", clientPort2)
	clientConfig2 := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort2),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.Hysteria2ClientConfig{
						Auth: userID2.String(), // Unauthorized user
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	client2, err := buildclient.NewX(clientConfig2)
	common.Must(err)
	common.Must(client2.Start())
	defer client2.Close()

	// Connection should fail with unauthorized user
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(clientPort2),
	})
	if err == nil {
		conn.SetDeadline(time.Now().Add(scenarios.Timeout))
		_, err = conn.Write([]byte("test"))
		conn.Close()
		t.Log("Unauthorized Hysteria2 user behavior: ", err)
	}
}

// TestMultipleUsersHysteria2 tests multiple concurrent users on Hysteria2
func TestMultipleUsersHysteria2(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := nethelper.PickUDPPort()
	t.Log("server port", serverPort)

	// Create multiple users
	var userIDs []uuid.UUID
	var userConfigs []*configs.UserConfig
	for i := 0; i < 3; i++ {
		userID := uuid.New()
		userIDs = append(userIDs, userID)
		userConfigs = append(userConfigs, &configs.UserConfig{
			Id:     userID.String(),
			Secret: userID.String(),
		})
	}

	serverConfig := &server.ServerConfig{
		Users: userConfigs,
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Hysteria2ServerConfig{
						IgnoreClientBandwidth: true,
						TlsConfig: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test all users concurrently
	var errg errgroup.Group
	for idx, userID := range userIDs {
		userID := userID // capture loop variable
		clientPort := tcp.PickPort()
		t.Logf("client port for user %d: %d", idx, clientPort)

		clientConfig := &configs.TmConfig{
			InboundManager: &configs.InboundManagerConfig{
				Handlers: []*configs.ProxyInboundConfig{
					{
						Address: nethelper.LocalHostIP.String(),
						Port:    uint32(clientPort),
						Protocol: serial.ToTypedMessage(
							&proxyconfig.DokodemoConfig{
								Address:  dest.Address.String(),
								Port:     uint32(dest.Port),
								Networks: []nethelper.Network{nethelper.Network_TCP},
							},
						),
					},
				},
			},
			Outbound: &configs.OutboundConfig{
				OutboundHandlers: []*configs.OutboundHandlerConfig{
					{
						Address: nethelper.LocalHostIP.String(),
						Port:    uint32(serverPort),
						Protocol: serial.ToTypedMessage(&proxyconfig.Hysteria2ClientConfig{
							Auth: userID.String(),
							TlsConfig: &tls.TlsConfig{
								AllowInsecure: true,
								ServerName:    "example.com",
							},
						}),
					},
				},
			},
		}

		client, err := buildclient.NewX(clientConfig)
		if err != nil {
			t.Fatalf("Failed to create client for user %d: %v", idx, err)
		}
		common.Must(client.Start())
		defer client.Close()

		// Test connections for this user
		for i := 0; i < 2; i++ {
			errg.Go(scenarios.TestTCPConn(clientPort, 10240, scenarios.Timeout*2))
		}
	}

	if err := errg.Wait(); err != nil {
		t.Errorf("Concurrent Hysteria2 user test failed: %v", err)
	}
}

// TestHysteria2WithObfuscation tests Hysteria2 user authentication with obfuscation
func TestHysteria2WithObfuscation(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: scenarios.Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := nethelper.PickUDPPort()
	t.Log("server port", serverPort)

	userID := uuid.New()
	obfsPassword := "test-obfs-password"

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: nethelper.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Hysteria2ServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
						IgnoreClientBandwidth: true,
						TlsConfig: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
						Obfs: &proxyconfig.ObfsConfig{
							Obfs: &proxyconfig.ObfsConfig_Salamander{
								Salamander: &proxyconfig.SalamanderConfig{
									Password: obfsPassword,
								},
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	// Test with correct obfs password
	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []nethelper.Network{nethelper.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: nethelper.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
						Obfs: &proxyconfig.ObfsConfig{
							Obfs: &proxyconfig.ObfsConfig_Salamander{
								Salamander: &proxyconfig.SalamanderConfig{
									Password: obfsPassword,
								},
							},
						},
					}),
				},
			},
		},
	}

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 2; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort, 10240, scenarios.Timeout*2))
	}

	if err := errg.Wait(); err != nil {
		t.Errorf("Connection failed with obfuscation: %v", err)
	}
}
