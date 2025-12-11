// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Portions of this code are modified from Xray-core/transport/internet/tls/ech.go
// Original source: https://github.com/XTLS/Xray-core

package tls

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/cryptobyte"
)

func (e *Engine) ApplyECH(config *tls.Config) {
	if len(e.config.EchConfig) > 0 {
		config.EncryptedClientHelloConfigList = e.config.EchConfig
		return
	}
	if e.config.EnableEch {
		if e.dnsServer == nil {
			log.Error().Msg("dns server is not set")
			return
		}
		if config.ServerName == "" {
			log.Error().Msg("server name is not set")
			return
		}
		echConfig, err := e.dnsServer.LookupECH(context.Background(),
			config.ServerName)
		if err != nil {
			log.Error().Msg("failed to lookup ECH config")
			return
		}
		config.EncryptedClientHelloConfigList = echConfig
	}
}

var ErrInvalidLen = errors.New("goech: invalid length")

func ConvertToGoECHKeys(data []byte) ([]tls.EncryptedClientHelloKey, error) {
	var keys []tls.EncryptedClientHelloKey
	s := cryptobyte.String(data)
	for !s.Empty() {
		if len(s) < 2 {
			return keys, ErrInvalidLen
		}
		keyLength := int(binary.BigEndian.Uint16(s[:2]))
		if len(s) < keyLength+4 {
			return keys, ErrInvalidLen
		}
		configLength := int(binary.BigEndian.Uint16(s[keyLength+2 : keyLength+4]))
		if len(s) < 2+keyLength+2+configLength {
			return keys, ErrInvalidLen
		}
		child := cryptobyte.String(s[:2+keyLength+2+configLength])
		var (
			sk, config cryptobyte.String
		)
		if !child.ReadUint16LengthPrefixed(&sk) || !child.ReadUint16LengthPrefixed(&config) || !child.Empty() {
			return keys, ErrInvalidLen
		}
		if !s.Skip(2 + keyLength + 2 + configLength) {
			return keys, ErrInvalidLen
		}
		keys = append(keys, tls.EncryptedClientHelloKey{
			Config:     config,
			PrivateKey: sk,
		})
	}
	return keys, nil
}
