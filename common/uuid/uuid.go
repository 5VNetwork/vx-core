package uuid

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/common"
)

var byteGroups = []int{8, 4, 4, 4, 12}
var emptyUUID = UUID{}

// NamespaceURL is the UUID namespace for URLs (RFC 4122)
var NamespaceURL = UUID{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

type UUID [16]byte

func (u UUID) ToSlice() []byte {
	return u.Bytes()
}

func (u UUID) IsSet() bool {
	return u != emptyUUID
}

// String returns the string representation of this UUID.
func (u UUID) String() string {
	bytes := u.Bytes()
	result := hex.EncodeToString(bytes[0 : byteGroups[0]/2])
	start := byteGroups[0] / 2
	for i := 1; i < len(byteGroups); i++ {
		nBytes := byteGroups[i] / 2
		result += "-"
		result += hex.EncodeToString(bytes[start : start+nBytes])
		start += nBytes
	}
	return result
}

// Bytes returns the bytes representation of this UUID.
func (u *UUID) Bytes() []byte {
	return u[:]
}

func (u *UUID) UID() *UUID {
	return u
}

// Equals returns true if this UUID equals another UUID by value.
func (u *UUID) Equals(another *UUID) bool {
	if u == nil && another == nil {
		return true
	}
	if u == nil || another == nil {
		return false
	}
	return bytes.Equal(u.Bytes(), another.Bytes())
}

// New creates a UUID with random value.
func New() UUID {
	var uuid UUID
	common.Must2(rand.Read(uuid.Bytes()))
	return uuid
}

// ParseBytes converts a UUID in byte form to object.
func ParseBytes(b []byte) (UUID, error) {
	var uuid UUID
	if len(b) != 16 {
		return uuid, fmt.Errorf("invalid UUID: %s", string(b))
	}
	copy(uuid[:], b)
	return uuid, nil
}

// ParseString converts a UUID in string form to object.
func ParseString(str string) (UUID, error) {
	var uuid UUID

	text := []byte(str)
	if len(text) < 32 {
		return uuid, errors.New("invalid UUID: " + str)
	}

	b := uuid.Bytes()

	for _, byteGroup := range byteGroups {
		if text[0] == '-' {
			text = text[1:]
		}

		if _, err := hex.Decode(b[:byteGroup/2], text[:byteGroup]); err != nil {
			return uuid, err
		}

		text = text[byteGroup:]
		b = b[byteGroup/2:]
	}

	return uuid, nil
}

// StringToUUID converts string to UUID. If str is not a valid UUID, it generates a Name-Based UUID v5.
func StringToUUID(str string) UUID {
	// Try to parse as a valid UUID first
	uuid, err := ParseString(str)
	if err == nil {
		return uuid
	}

	// If not a valid UUID, generate a Name-Based UUID v5 (SHA-1 based)
	return NewV5(NamespaceURL, str)
}

// NewV5 generates a UUID v5 (Name-Based UUID using SHA-1) from a namespace and name.
func NewV5(namespace UUID, name string) UUID {
	var uuid UUID

	// Create SHA-1 hash of namespace + name
	hash := sha1.New()
	hash.Write(namespace[:])
	hash.Write([]byte(name))
	sum := hash.Sum(nil)

	// Copy first 16 bytes of hash to UUID
	copy(uuid[:], sum[:16])

	// Set version to 5 (Name-Based UUID with SHA-1)
	// Version bits: 0101 in bits 12-15 of time_hi_and_version field (byte 6)
	uuid[6] = (uuid[6] & 0x0f) | 0x50

	// Set variant to RFC 4122 (bits 10xx in clock_seq_hi_and_reserved field, byte 8)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid
}
