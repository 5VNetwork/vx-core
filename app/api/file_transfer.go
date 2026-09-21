// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	createinbound "github.com/5vnetwork/vx-core/app/create/inbound"
	createoutbound "github.com/5vnetwork/vx-core/app/create/outbound"
	"github.com/5vnetwork/vx-core/app/filetransfer"
	proxyinbound "github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/util/uri"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	fileTransferBandwidthMbps = 1000
	realmReadyTimeout         = 30 * time.Second
)

func (a *Api) StartFileReceive(req *StartFileReceiveRequest, stream Api_StartFileReceiveServer) error {
	if req.GetInbound() == nil {
		return status.Error(codes.InvalidArgument, "inbound is required")
	}
	saveDir := req.GetSaveDir()
	if saveDir == "" {
		return status.Error(codes.InvalidArgument, "save_dir is required")
	}
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return status.Errorf(codes.InvalidArgument, "save_dir: %v", err)
	}

	ctx, sess, err := a.beginFileSession(true, stream.Context())
	if err != nil {
		return err
	}
	defer a.endFileSession(true, sess)

	if err := stream.Send(&FileTransferEvent{State: FileTransferEvent_STATE_STARTING}); err != nil {
		return err
	}

	inboundCfg := proto.Clone(req.GetInbound()).(*configs.ProxyInboundConfig)
	tweakHysteriaServer(inboundCfg)
	if inboundCfg.Tag == "" {
		inboundCfg.Tag = "fileReceive"
	}

	events := make(chan *FileTransferEvent, 32)
	receiver := &filetransfer.Receiver{
		SaveDir: saveDir,
		OnEvent: func(ev filetransfer.ReceiveEvent) {
			out := receiveEventToProto(ev)
			select {
			case events <- out:
			case <-ctx.Done():
			}
		},
	}

	shareURL, err := fileReceiveShareURL(inboundCfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to build file receive share url")
	}

	in, err := createinbound.NewInbound(inboundCfg, receiver, policy.New(), a.getIPResolver(), a.getDialerFactory())
	if err != nil {
		return status.Errorf(codes.Internal, "create inbound: %v", err)
	}
	if err := in.Start(); err != nil {
		_ = in.Close()
		return status.Errorf(codes.Internal, "start inbound: %v", err)
	}
	defer in.Close()

	if inboundHasRealm(inboundCfg) {
		if err := waitRealmReady(ctx, in, realmReadyTimeout); err != nil {
			_ = stream.Send(&FileTransferEvent{
				State:    FileTransferEvent_STATE_FAILED,
				ShareUrl: shareURL,
				Error:    err.Error(),
			})
			return nil
		}
	}

	if err := stream.Send(&FileTransferEvent{
		State:    FileTransferEvent_STATE_WAITING,
		ShareUrl: shareURL,
	}); err != nil {
		return err
	}

	for {
		select {
		case ev := <-events:
			if ev.ShareUrl == "" {
				ev.ShareUrl = shareURL
			}
			if err := stream.Send(ev); err != nil {
				return err
			}
		case <-ctx.Done():
			_ = stream.Send(&FileTransferEvent{
				State:    FileTransferEvent_STATE_CANCELLED,
				ShareUrl: shareURL,
			})
			return nil
		}
	}
}

func (a *Api) StartFileSend(req *StartFileSendRequest, stream Api_StartFileSendServer) error {
	if req.GetRealmUrl() == "" {
		return status.Error(codes.InvalidArgument, "realm_url is required")
	}
	filePath := req.GetFilePath()
	if filePath == "" {
		return status.Error(codes.InvalidArgument, "file_path is required")
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "file_path: %v", err)
	}
	if info.IsDir() {
		return status.Error(codes.InvalidArgument, "file_path is a directory")
	}

	ctx, sess, err := a.beginFileSession(false, stream.Context())
	if err != nil {
		return err
	}
	defer a.endFileSession(false, sess)

	filename := filepath.Base(filePath)
	total := uint64(info.Size())
	send := func(ev *FileTransferEvent) error {
		if ev.Filename == "" {
			ev.Filename = filename
		}
		if ev.BytesTotal == 0 {
			ev.BytesTotal = total
		}
		return stream.Send(ev)
	}

	if err := send(&FileTransferEvent{State: FileTransferEvent_STATE_CONNECTING}); err != nil {
		return err
	}

	decoded, err := util.Decode(req.GetRealmUrl(), nil)
	if err != nil {
		_ = send(&FileTransferEvent{State: FileTransferEvent_STATE_FAILED, Error: err.Error()})
		return nil
	}
	if len(decoded.Configs) == 0 {
		_ = send(&FileTransferEvent{State: FileTransferEvent_STATE_FAILED, Error: "failed to decode realm url"})
		return nil
	}
	outCfg := proto.Clone(decoded.Configs[0]).(*configs.OutboundHandlerConfig)
	tweakHysteriaClient(outCfg)

	handler, err := createoutbound.NewHandler(&createoutbound.HandlerConfig{
		HandlerConfig: &configs.HandlerConfig{
			Type: &configs.HandlerConfig_Outbound{Outbound: outCfg},
		},
		DialerFactory:               a.getDialerFactory(),
		Policy:                      policy.New(),
		IPResolver:                  a.getIPResolver(),
		EchResolver:                 a.echResolver,
		IPResolverForRequestAddress: a.getIPResolver(),
	})
	if err != nil {
		_ = send(&FileTransferEvent{State: FileTransferEvent_STATE_FAILED, Error: err.Error()})
		return nil
	}
	defer common.Close(handler)

	err = filetransfer.Send(ctx, handler, filePath, func(done, tot uint64) {
		state := FileTransferEvent_STATE_TRANSFERRING
		if tot > 0 && done >= tot {
			state = FileTransferEvent_STATE_COMPLETED
		}
		_ = send(&FileTransferEvent{
			State:      state,
			Filename:   filename,
			BytesDone:  done,
			BytesTotal: tot,
		})
	})
	if err != nil {
		if ctx.Err() != nil {
			_ = send(&FileTransferEvent{State: FileTransferEvent_STATE_CANCELLED, Error: err.Error()})
			return nil
		}
		_ = send(&FileTransferEvent{State: FileTransferEvent_STATE_FAILED, Error: err.Error()})
		return nil
	}
	_ = send(&FileTransferEvent{
		State:      FileTransferEvent_STATE_COMPLETED,
		Filename:   filename,
		BytesDone:  total,
		BytesTotal: total,
	})
	return nil
}

type fileSession struct {
	cancel context.CancelFunc
}

func (a *Api) beginFileSession(receive bool, parent context.Context) (context.Context, *fileSession, error) {
	ctx, cancel := context.WithCancel(parent)
	sess := &fileSession{cancel: cancel}

	a.ftLock.Lock()
	var prev *fileSession
	if receive {
		prev = a.receiveSession
		a.receiveSession = sess
	} else {
		prev = a.sendSession
		a.sendSession = sess
	}
	a.ftLock.Unlock()

	if prev != nil {
		prev.cancel()
	}
	return ctx, sess, nil
}

func (a *Api) endFileSession(receive bool, sess *fileSession) {
	if sess != nil {
		sess.cancel()
	}
	a.ftLock.Lock()
	defer a.ftLock.Unlock()
	if receive {
		if a.receiveSession == sess {
			a.receiveSession = nil
		}
		return
	}
	if a.sendSession == sess {
		a.sendSession = nil
	}
}

func receiveEventToProto(ev filetransfer.ReceiveEvent) *FileTransferEvent {
	out := &FileTransferEvent{
		Filename:   ev.Filename,
		SavePath:   ev.SavePath,
		BytesDone:  ev.Done,
		BytesTotal: ev.Total,
	}
	switch {
	case ev.Err != nil:
		out.State = FileTransferEvent_STATE_FAILED
		out.Error = ev.Err.Error()
	case ev.Complete:
		out.State = FileTransferEvent_STATE_COMPLETED
	default:
		out.State = FileTransferEvent_STATE_TRANSFERRING
	}
	return out
}

func fileReceiveShareURL(inboundCfg *configs.ProxyInboundConfig) (string, error) {
	outs, err := util.InboundConfigToOutboundConfig("FileTransfer", inboundCfg, "")
	if err != nil {
		return "", err
	}
	if len(outs) == 0 {
		return "", errors.New("no outbound config")
	}
	tweakHysteriaClient(outs[0])
	return uri.ToUrl(outs[0])
}

func inboundHasRealm(in *configs.ProxyInboundConfig) bool {
	if in == nil {
		return false
	}
	check := func(a *anypb.Any) bool {
		if a == nil {
			return false
		}
		msg, err := a.UnmarshalNew()
		if err != nil {
			return false
		}
		h, ok := msg.(*configs.Hysteria2ServerConfig)
		return ok && h.GetRealm().GetRealmAddr() != ""
	}
	if check(in.Protocol) {
		return true
	}
	for _, p := range in.Protocols {
		if check(p) {
			return true
		}
	}
	return false
}

func tweakHysteriaServer(in *configs.ProxyInboundConfig) {
	if in == nil {
		return
	}
	tweak := func(a *anypb.Any) *anypb.Any {
		if a == nil {
			return a
		}
		msg, err := a.UnmarshalNew()
		if err != nil {
			return a
		}
		h, ok := msg.(*configs.Hysteria2ServerConfig)
		if !ok {
			return a
		}
		h.IgnoreClientBandwidth = true
		if h.Bandwidth == nil {
			h.Bandwidth = &configs.BandwidthConfig{}
		}
		if h.Bandwidth.MaxRx == 0 {
			h.Bandwidth.MaxRx = fileTransferBandwidthMbps
		}
		if h.Bandwidth.MaxTx == 0 {
			h.Bandwidth.MaxTx = fileTransferBandwidthMbps
		}
		return serial.ToTypedMessage(h)
	}
	in.Protocol = tweak(in.Protocol)
	for i := range in.Protocols {
		in.Protocols[i] = tweak(in.Protocols[i])
	}
}

func tweakHysteriaClient(out *configs.OutboundHandlerConfig) {
	if out == nil || out.Protocol == nil {
		return
	}
	msg, err := out.Protocol.UnmarshalNew()
	if err != nil {
		return
	}
	h, ok := msg.(*configs.Hysteria2ClientConfig)
	if !ok {
		return
	}
	if h.Bandwidth == nil {
		h.Bandwidth = &configs.BandwidthConfig{}
	}
	h.Bandwidth.MaxTx = fileTransferBandwidthMbps
	h.Bandwidth.MaxRx = fileTransferBandwidthMbps
	out.Protocol = serial.ToTypedMessage(h)
}

func waitRealmReady(ctx context.Context, in proxyinbound.Inbound, timeout time.Duration) error {
	hy := hysteriaInboundFrom(in)
	if hy == nil {
		return errors.New("hysteria inbound not found")
	}
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		active, _, _, _ := hy.RealmStatus()
		if active {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for realm registration")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func hysteriaInboundFrom(in proxyinbound.Inbound) *hysteria2.Inbound {
	p, ok := in.(*proxyinbound.ProxyInbound)
	if !ok {
		return nil
	}
	for _, w := range p.GetWorkers() {
		if h, ok := w.(*hysteria2.Inbound); ok {
			return h
		}
	}
	return nil
}
