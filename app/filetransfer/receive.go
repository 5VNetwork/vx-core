// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
)

// ReceiveEvent is emitted by Receiver as files arrive.
type ReceiveEvent struct {
	Filename string
	SavePath string
	Done     uint64
	Total    uint64
	Complete bool
	Err      error
}

// Receiver writes inbound file-transfer streams to SaveDir.
type Receiver struct {
	SaveDir  string
	OnEvent  func(ReceiveEvent)
	onPacket i.PacketHandler
}

func (r *Receiver) emit(ev ReceiveEvent) {
	if r != nil && r.OnEvent != nil {
		r.OnEvent(ev)
	}
}

func (r *Receiver) HandleFlow(ctx context.Context, _ net.Destination, rw buf.ReaderWriter) error {
	ev, err := r.receive(ctx, rw)
	if err != nil {
		ev.Err = err
		r.emit(ev)
		return err
	}
	ev.Complete = true
	r.emit(ev)
	_ = rw.CloseWrite()
	return nil
}

func (r *Receiver) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	return fmt.Errorf("file transfer does not accept UDP")
}

func (r *Receiver) receive(ctx context.Context, rw buf.Reader) (ev ReceiveEvent, err error) {
	br := &buf.BufferedReader{Reader: rw}
	header, err := decodeHeader(br)
	if err != nil {
		return ReceiveEvent{}, err
	}
	name, err := sanitizeFilename(header.Name)
	if err != nil {
		return ReceiveEvent{}, err
	}
	savePath := uniquePath(r.SaveDir, name)
	ev = ReceiveEvent{Filename: name, SavePath: savePath, Total: header.Size}
	r.emit(ev)

	if err = os.MkdirAll(r.SaveDir, 0o755); err != nil {
		return ev, fmt.Errorf("create save dir: %w", err)
	}
	tmpPath := savePath + ".vxft.tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return ev, fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	progress := newThrottledProgress(header.Size, func(done, total uint64) {
		r.emit(ReceiveEvent{
			Filename: name,
			SavePath: savePath,
			Done:     done,
			Total:    total,
		})
	})
	writer := buf.NewWriter(f)
	limited := &limitedReader{r: br, remain: int64(header.Size)}
	if copyErr := buf.Copy(buf.NewReader(limited), writer, progressDataHandler{p: progress}); copyErr != nil {
		err = fmt.Errorf("write file: %w", copyErr)
		return ev, err
	}
	if limited.remain > 0 {
		err = fmt.Errorf("short file: missing %d bytes", limited.remain)
		return ev, err
	}
	if cerr := f.Close(); cerr != nil {
		err = cerr
		return ev, err
	}
	if rerr := os.Rename(tmpPath, savePath); rerr != nil {
		err = fmt.Errorf("rename: %w", rerr)
		return ev, err
	}
	progress.Force(header.Size)
	ev.Done = header.Size
	ev.SavePath = savePath
	if ctx.Err() != nil {
		err = ctx.Err()
		_ = os.Remove(savePath)
		return ev, err
	}
	return ev, nil
}

type limitedReader struct {
	r      io.Reader
	remain int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	if l.remain <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.remain {
		p = p[:l.remain]
	}
	n, err := l.r.Read(p)
	l.remain -= int64(n)
	return n, err
}
