package util

/*
Curve25519Genkey is copied from Xray-core project,
source code: https://github.com/XTLS/Xray-core
*/

import (
	"encoding/base64"
	"errors"
	"math/rand"

	"golang.org/x/crypto/curve25519"
)

func Curve25519Genkey(StdEncoding bool, input_base64 string) (string, string, error) {
	var err error
	var privateKey, publicKey []byte
	var encoding *base64.Encoding
	if StdEncoding {
		encoding = base64.StdEncoding
	} else {
		encoding = base64.RawURLEncoding
	}

	if len(input_base64) > 0 {
		privateKey, err = encoding.DecodeString(input_base64)
		if err != nil {
			return "", "", err
		}
		if len(privateKey) != curve25519.ScalarSize {
			return "", "", errors.New("invalid length of private key")
		}
	}

	if privateKey == nil {
		privateKey = make([]byte, curve25519.ScalarSize)
		if _, err = rand.Read(privateKey); err != nil {
			return "", "", err
		}
	}

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html.
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	if publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint); err != nil {
		return "", "", err
	}

	return encoding.EncodeToString(publicKey), encoding.EncodeToString(privateKey), nil
}
