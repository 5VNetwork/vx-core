// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
)

func TestSanitizeFilename(t *testing.T) {
	ok, err := sanitizeFilename("hello.txt")
	if err != nil || ok != "hello.txt" {
		t.Fatalf("got %q err %v", ok, err)
	}
	ok, err = sanitizeFilename(`C:\tmp\..\secret.txt`)
	if err != nil || ok != "secret.txt" {
		t.Fatalf("got %q err %v", ok, err)
	}
	if _, err := sanitizeFilename(".."); err == nil {
		t.Fatal("expected error")
	}
	if _, err := sanitizeFilename(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := sanitizeFilename("   "); err == nil {
		t.Fatal("expected error")
	}
}

func TestUniquePath(t *testing.T) {
	dir := t.TempDir()
	name := "a.txt"
	p1 := uniquePath(dir, name)
	if err := os.WriteFile(p1, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	p2 := uniquePath(dir, name)
	if p2 == p1 {
		t.Fatal("expected uniquified path")
	}
	if filepath.Base(p2) != "a (1).txt" {
		t.Fatalf("got %s", filepath.Base(p2))
	}
}

func TestFrameRoundTrip(t *testing.T) {
	encoded, err := encodeHeader(Header{Name: "photo.png", Size: 12345})
	if err != nil {
		t.Fatal(err)
	}
	h, err := decodeHeader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if h.Name != "photo.png" || h.Size != 12345 {
		t.Fatalf("got %+v", h)
	}
}

func TestWriteAndReceiveFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	payload := bytes.Repeat([]byte("vx-file-transfer"), 1024)
	if err := os.WriteFile(src, payload, 0o644); err != nil {
		t.Fatal(err)
	}

	client, server := pipe.NewLinks(-1, false)
	recvDir := filepath.Join(dir, "dl")
	var got ReceiveEvent
	receiver := &Receiver{
		SaveDir: recvDir,
		OnEvent: func(ev ReceiveEvent) { got = ev },
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- receiver.HandleFlow(context.Background(), net.Destination{}, server)
	}()

	if err := WriteFile(context.Background(), client, src, nil); err != nil {
		t.Fatal(err)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	if !got.Complete || got.Filename != "src.bin" {
		t.Fatalf("event %+v", got)
	}
	data, err := os.ReadFile(got.SavePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("payload mismatch %d vs %d", len(data), len(payload))
	}
}

func TestRejectPathTraversalName(t *testing.T) {
	if _, err := encodeHeader(Header{Name: "", Size: 1}); err == nil {
		t.Fatal("expected error")
	}
}
