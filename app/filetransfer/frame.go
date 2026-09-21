// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	magic          = "VXFT"
	version   byte = 1
	maxNameLen     = 1024
)

var (
	errBadMagic   = errors.New("invalid file transfer magic")
	errBadVersion = errors.New("unsupported file transfer version")
	errNameLen    = errors.New("invalid filename length")
)

// Header is the metadata sent ahead of the file payload.
type Header struct {
	Name string
	Size uint64
}

func encodeHeader(h Header) ([]byte, error) {
	name := []byte(h.Name)
	if len(name) == 0 || len(name) > maxNameLen {
		return nil, errNameLen
	}
	buf := make([]byte, 4+1+2+len(name)+8)
	copy(buf[0:4], magic)
	buf[4] = version
	binary.BigEndian.PutUint16(buf[5:7], uint16(len(name)))
	copy(buf[7:7+len(name)], name)
	binary.BigEndian.PutUint64(buf[7+len(name):], h.Size)
	return buf, nil
}

func decodeHeader(r io.Reader) (Header, error) {
	var prefix [7]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return Header{}, fmt.Errorf("read header prefix: %w", err)
	}
	if string(prefix[0:4]) != magic {
		return Header{}, errBadMagic
	}
	if prefix[4] != version {
		return Header{}, fmt.Errorf("%w: %d", errBadVersion, prefix[4])
	}
	nameLen := int(binary.BigEndian.Uint16(prefix[5:7]))
	if nameLen <= 0 || nameLen > maxNameLen {
		return Header{}, errNameLen
	}
	nameBuf := make([]byte, nameLen)
	if _, err := io.ReadFull(r, nameBuf); err != nil {
		return Header{}, fmt.Errorf("read filename: %w", err)
	}
	var sizeBuf [8]byte
	if _, err := io.ReadFull(r, sizeBuf[:]); err != nil {
		return Header{}, fmt.Errorf("read size: %w", err)
	}
	return Header{
		Name: string(nameBuf),
		Size: binary.BigEndian.Uint64(sizeBuf[:]),
	}, nil
}
