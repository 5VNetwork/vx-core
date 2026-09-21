// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	createinbound "github.com/5vnetwork/vx-core/app/create/inbound"
	createoutbound "github.com/5vnetwork/vx-core/app/create/outbound"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

func TestHysteriaTCPFileRoundTrip(t *testing.T) {
	port := net.PickUDPPort()
	secret := uuid.New().String()
	dir := t.TempDir()
	src := filepath.Join(dir, "payload.bin")
	payload := bytes.Repeat([]byte("hysteria-file-transfer"), 2048)
	if err := os.WriteFile(src, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	recvDir := filepath.Join(dir, "dl")

	done := make(chan ReceiveEvent, 1)
	receiver := &Receiver{
		SaveDir: recvDir,
		OnEvent: func(ev ReceiveEvent) {
			if ev.Complete || ev.Err != nil {
				select {
				case done <- ev:
				default:
				}
			}
		},
	}

	inboundCfg := &configs.ProxyInboundConfig{
		Tag:     "fileReceive",
		Address: "127.0.0.1",
		Port:    uint32(port),
		Users:   []*configs.UserConfig{{Id: "u", Secret: secret}},
		Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
			IgnoreClientBandwidth: true,
			TlsConfig: &tls.TlsConfig{
				Certificates: []*tls.Certificate{
					tls.ParseCertificate(cert.MustGenerate(nil)),
				},
			},
		}),
	}
	in, err := createinbound.NewInbound(
		inboundCfg,
		receiver,
		policy.New(),
		&dns.GoDnsResolver{},
		transport.DefaultDialerFactory(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := in.Start(); err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	client, err := createoutbound.NewOutHandler(&createoutbound.Config{
		OutboundHandlerConfig: &configs.OutboundHandlerConfig{
			Tag:     "fileSend",
			Address: "127.0.0.1",
			Port:    uint32(port),
			Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
				Auth: secret,
				Bandwidth: &configs.BandwidthConfig{
					MaxTx: 100,
					MaxRx: 100,
				},
				TlsConfig: &tls.TlsConfig{
					AllowInsecure: true,
					ServerName:    "example.com",
				},
			}),
		},
		DialerFactory: transport.DefaultDialerFactory(),
		Policy:        policy.New(),
		IPResolver:    &dns.GoDnsResolver{},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := Send(ctx, client, src, nil); err != nil {
		t.Fatal(err)
	}

	select {
	case ev := <-done:
		if ev.Err != nil {
			t.Fatal(ev.Err)
		}
		got, err := os.ReadFile(ev.SavePath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("payload mismatch %d vs %d", len(got), len(payload))
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for receive")
	}
}
