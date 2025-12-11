// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This code are modified from Xray-core/main/commands/all/tls/ech.go
// Original source: https://github.com/XTLS/Xray-core

package util

import (
	"crypto/ecdh"
	"crypto/rand"
	"io"

	"github.com/5vnetwork/vx-core/common"
	"github.com/xtls/reality"
	"github.com/xtls/reality/hpke"
	"golang.org/x/crypto/cryptobyte"
)

func ExecuteECH(serverName string) ([]byte, []byte, error) {
	var kem uint16

	// if *input_pqSignatureSchemesEnabled {
	// 	kem = 0x30 // hpke.KEM_X25519_KYBER768_DRAFT00
	// } else {
	kem = hpke.DHKEM_X25519_HKDF_SHA256
	// }

	echConfig, priv, err := GenerateECHKeySet(0, serverName, kem)
	common.Must(err)

	var configBuffer, keyBuffer []byte
	configBytes, _ := MarshalBinary(echConfig)
	var b cryptobyte.Builder
	b.AddUint16LengthPrefixed(func(child *cryptobyte.Builder) {
		child.AddBytes(configBytes)
	})
	configBuffer, _ = b.Bytes()
	var b2 cryptobyte.Builder
	b2.AddUint16(uint16(len(priv)))
	b2.AddBytes(priv)
	b2.AddUint16(uint16(len(configBytes)))
	b2.AddBytes(configBytes)
	keyBuffer, _ = b2.Bytes()

	return configBuffer, keyBuffer, nil
}

const ExtensionEncryptedClientHello = 0xfe0d
const KDF_HKDF_SHA384 = 0x0002
const KDF_HKDF_SHA512 = 0x0003

func GenerateECHKeySet(configID uint8, domain string, kem uint16) (reality.EchConfig, []byte, error) {
	config := reality.EchConfig{
		Version:    ExtensionEncryptedClientHello,
		ConfigID:   configID,
		PublicName: []byte(domain),
		KemID:      kem,
		SymmetricCipherSuite: []reality.EchCipher{
			{KDFID: hpke.KDF_HKDF_SHA256, AEADID: hpke.AEAD_AES_128_GCM},
			{KDFID: hpke.KDF_HKDF_SHA256, AEADID: hpke.AEAD_AES_256_GCM},
			{KDFID: hpke.KDF_HKDF_SHA256, AEADID: hpke.AEAD_ChaCha20Poly1305},
			{KDFID: KDF_HKDF_SHA384, AEADID: hpke.AEAD_AES_128_GCM},
			{KDFID: KDF_HKDF_SHA384, AEADID: hpke.AEAD_AES_256_GCM},
			{KDFID: KDF_HKDF_SHA384, AEADID: hpke.AEAD_ChaCha20Poly1305},
			{KDFID: KDF_HKDF_SHA512, AEADID: hpke.AEAD_AES_128_GCM},
			{KDFID: KDF_HKDF_SHA512, AEADID: hpke.AEAD_AES_256_GCM},
			{KDFID: KDF_HKDF_SHA512, AEADID: hpke.AEAD_ChaCha20Poly1305},
		},
		MaxNameLength: 0,
		Extensions:    nil,
	}
	// if kem == hpke.DHKEM_X25519_HKDF_SHA256 {
	curve := ecdh.X25519()
	priv := make([]byte, 32) //x25519
	_, err := io.ReadFull(rand.Reader, priv)
	if err != nil {
		return config, nil, err
	}
	privKey, _ := curve.NewPrivateKey(priv)
	config.PublicKey = privKey.PublicKey().Bytes()
	return config, priv, nil
	// }
	// TODO: add mlkem768 (former kyber768 draft00). The golang mlkem private key is 64 bytes seed?
}

// reference github.com/OmarTariq612/goech
func MarshalBinary(ech reality.EchConfig) ([]byte, error) {
	var b cryptobyte.Builder
	b.AddUint16(ech.Version)
	b.AddUint16LengthPrefixed(func(child *cryptobyte.Builder) {
		child.AddUint8(ech.ConfigID)
		child.AddUint16(ech.KemID)
		child.AddUint16(uint16(len(ech.PublicKey)))
		child.AddBytes(ech.PublicKey)
		child.AddUint16LengthPrefixed(func(child *cryptobyte.Builder) {
			for _, cipherSuite := range ech.SymmetricCipherSuite {
				child.AddUint16(cipherSuite.KDFID)
				child.AddUint16(cipherSuite.AEADID)
			}
		})
		child.AddUint8(ech.MaxNameLength)
		child.AddUint8(uint8(len(ech.PublicName)))
		child.AddBytes(ech.PublicName)
		child.AddUint16LengthPrefixed(func(child *cryptobyte.Builder) {
			for _, extention := range ech.Extensions {
				child.AddUint16(extention.Type)
				child.AddBytes(extention.Data)
			}
		})
	})
	return b.Bytes()
}
