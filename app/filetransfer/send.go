// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
)

const FileDestHost = "file.vx"

func fileDestination() net.Destination {
	return net.TCPDestination(net.DomainAddress(FileDestHost), 1)
}

// Send copies filePath over handler as a file-transfer stream.
func Send(ctx context.Context, handler i.Handler, filePath string, progress ProgressFunc) error {
	dest := fileDestination()
	var rw buf.ReaderWriter
	var closer interface{ Close() error }

	if d, ok := handler.(i.ProxyDialer); ok {
		fc, err := d.ProxyDial(ctx, dest, nil)
		if err != nil {
			return fmt.Errorf("dial file transfer: %w", err)
		}
		rw = fc
		closer = fc
	} else {
		conn, err := (&util.FlowHandlerToDialer{FlowHandler: handler}).Dial(ctx, dest)
		if err != nil {
			return fmt.Errorf("dial file transfer: %w", err)
		}
		rw = buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn)
		closer = conn
	}
	defer closer.Close()
	stop := context.AfterFunc(ctx, func() { _ = closer.Close() })
	defer stop()
	return WriteFile(ctx, rw, filePath, progress)
}

// WriteFile writes a framed file to rw.
func WriteFile(ctx context.Context, rw buf.ReaderWriter, filePath string, progress ProgressFunc) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	name, err := sanitizeFilename(filepath.Base(filePath))
	if err != nil {
		return err
	}
	header, err := encodeHeader(Header{Name: name, Size: uint64(info.Size())})
	if err != nil {
		return err
	}
	if err := rw.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes(header)}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	throttled := newThrottledProgress(uint64(info.Size()), progress)
	if err := buf.Copy(ctxReader{ctx: ctx, r: buf.NewReader(f)}, rw, progressDataHandler{p: throttled}); err != nil {
		return fmt.Errorf("send file: %w", err)
	}
	throttled.Force(uint64(info.Size()))
	_ = rw.CloseWrite()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

type ctxReader struct {
	ctx context.Context
	r   buf.Reader
}

func (c ctxReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if err := c.ctx.Err(); err != nil {
		return nil, err
	}
	return c.r.ReadMultiBuffer()
}
