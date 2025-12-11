package shadowsocks

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/common/antireplay"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/crypto"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/protocol"
)

type CipherType int32

const (
	CipherType_AES_128_GCM       CipherType = 0
	CipherType_AES_256_GCM       CipherType = 1
	CipherType_CHACHA20_POLY1305 CipherType = 2
	CipherType_NONE              CipherType = 3
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct {
	Cipher           Cipher
	Key              []byte
	Uid              string
	ReplayFilter     antireplay.GeneralizedReplayFilter
	ReducedIVEntropy bool
}

func NewMemoryAccount(uid string, cipher CipherType, password string,
	reducedIVEntropy, ivCheck bool) (*MemoryAccount, error) {
	Cipher, err := getCipher(cipher)
	if err != nil {
		return nil, fmt.Errorf("failed to get cipher: %w", err)
	}
	m := &MemoryAccount{
		Uid:    uid,
		Cipher: Cipher,
		Key:    passwordToCipherKey(password, Cipher.KeySize()),
		ReplayFilter: func() antireplay.GeneralizedReplayFilter {
			if ivCheck {
				return antireplay.NewBloomRing()
			}
			return nil
		}(),
		ReducedIVEntropy: reducedIVEntropy,
	}
	return m, nil
}

func (a *MemoryAccount) resetUser(uid, password string) {
	a.Uid = uid
	a.Key = passwordToCipherKey(password, a.Cipher.KeySize())
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(another protocol.Account) bool {
	if account, ok := another.(*MemoryAccount); ok {
		return bytes.Equal(a.Key, account.Key)
	}
	return false
}

func (a *MemoryAccount) CheckIV(iv []byte) error {
	if a.ReplayFilter == nil {
		return nil
	}
	if a.ReplayFilter.Check(iv) {
		return nil
	}
	return errors.New("IV is not unique")
}

func passwordToCipherKey(password string, keySize int32) []byte {
	key := make([]byte, 0, keySize)

	md5Sum := md5.Sum([]byte(password))
	key = append(key, md5Sum[:]...)

	for int32(len(key)) < keySize {
		md5Hash := md5.New()
		common.Must2(md5Hash.Write(md5Sum[:]))
		common.Must2(md5Hash.Write([]byte(password)))
		md5Hash.Sum(md5Sum[:0])

		key = append(key, md5Sum[:]...)
	}
	return key
}

func getCipher(a CipherType) (Cipher, error) {
	switch a {
	case CipherType_AES_128_GCM:
		return &AEADCipher{
			KeyBytes:        16,
			IVBytes:         16,
			AEADAuthCreator: createAesGcm,
		}, nil
	case CipherType_AES_256_GCM:
		return &AEADCipher{
			KeyBytes:        32,
			IVBytes:         32,
			AEADAuthCreator: createAesGcm,
		}, nil
	case CipherType_CHACHA20_POLY1305:
		return &AEADCipher{
			KeyBytes:        32,
			IVBytes:         32,
			AEADAuthCreator: createChaCha20Poly1305,
		}, nil
	case CipherType_NONE:
		return NoneCipher{}, nil
	default:
		return nil, errors.New("Unsupported cipher.")
	}
}

func createAesGcm(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	common.Must(err)
	gcm, err := cipher.NewGCM(block)
	common.Must(err)
	return gcm
}

func createChaCha20Poly1305(key []byte) cipher.AEAD {
	ChaChaPoly1305, err := chacha20poly1305.New(key)
	common.Must(err)
	return ChaChaPoly1305
}

// Cipher is an interface for all Shadowsocks ciphers.
type Cipher interface {
	KeySize() int32
	IVSize() int32
	NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error)
	NewEncryptionWriterIO(key []byte, iv []byte, writer io.Writer) (io.Writer, error)
	NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error)
	NewDecryptionReaderIO(key []byte, iv []byte, reader io.Reader) (io.Reader, error)
	IsAEAD() bool
	EncodePacket(key []byte, b *buf.Buffer) error
	DecodePacket(key []byte, b *buf.Buffer) error
}

type AEADCipher struct {
	KeyBytes        int32
	IVBytes         int32
	AEADAuthCreator func(key []byte) cipher.AEAD
}

func (*AEADCipher) IsAEAD() bool {
	return true
}

func (c *AEADCipher) KeySize() int32 {
	return c.KeyBytes
}

func (c *AEADCipher) IVSize() int32 {
	return c.IVBytes
}

func (c *AEADCipher) createAuthenticator(key []byte, iv []byte) *crypto.AEADAuthenticator {
	nonce := crypto.GenerateInitialAEADNonce()
	subkey := make([]byte, c.KeyBytes)
	hkdfSHA1(key, iv, subkey)
	return &crypto.AEADAuthenticator{
		AEAD:           c.AEADAuthCreator(subkey),
		NonceGenerator: nonce,
	}
}

func (c *AEADCipher) NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationWriter(auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, writer, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) NewEncryptionWriterIO(key []byte, iv []byte, writer io.Writer) (io.Writer, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationWriterIO(auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, writer, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationReader(context.Background(), auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, reader, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) NewDecryptionReaderIO(key []byte, iv []byte, reader io.Reader) (io.Reader, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationReader1(auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, reader, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) EncodePacket(key []byte, b *buf.Buffer) error {
	ivLen := c.IVSize()
	payloadLen := b.Len()
	auth := c.createAuthenticator(key, b.BytesTo(ivLen))

	b.Extend(int32(auth.Overhead()))
	_, err := auth.Seal(b.BytesTo(ivLen), b.BytesRange(ivLen, payloadLen))
	return err
}

func (c *AEADCipher) DecodePacket(key []byte, b *buf.Buffer) error {
	if b.Len() <= c.IVSize() {
		return errors.New("insufficient data: ", b.Len())
	}
	ivLen := c.IVSize()
	payloadLen := b.Len()
	auth := c.createAuthenticator(key, b.BytesTo(ivLen))

	bbb, err := auth.Open(b.BytesTo(ivLen), b.BytesRange(ivLen, payloadLen))
	if err != nil {
		return err
	}
	b.Resize(ivLen, int32(len(bbb)))
	return nil
}

type NoneCipher struct{}

func (NoneCipher) KeySize() int32 { return 0 }
func (NoneCipher) IVSize() int32  { return 0 }
func (NoneCipher) IsAEAD() bool {
	return false
}

func (NoneCipher) NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error) {
	return buf.NewReader(reader), nil
}

func (NoneCipher) NewDecryptionReaderIO(key []byte, iv []byte, reader io.Reader) (io.Reader, error) {
	return reader, nil
}

func (NoneCipher) NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error) {
	return buf.NewWriter(writer), nil
}

func (NoneCipher) NewEncryptionWriterIO(key []byte, iv []byte, writer io.Writer) (io.Writer, error) {
	return writer, nil
}

func (NoneCipher) EncodePacket(key []byte, b *buf.Buffer) error {
	return nil
}

func (NoneCipher) DecodePacket(key []byte, b *buf.Buffer) error {
	return nil
}

// func CipherFromString(c string) CipherType {
// 	switch strings.ToLower(c) {
// 	case "aes-128-gcm", "aes_128_gcm", "aead_aes_128_gcm":
// 		return CipherType_AES_128_GCM
// 	case "aes-256-gcm", "aes_256_gcm", "aead_aes_256_gcm":
// 		return CipherType_AES_256_GCM
// 	case "chacha20-poly1305", "chacha20_poly1305", "aead_chacha20_poly1305", "chacha20-ietf-poly1305":
// 		return CipherType_CHACHA20_POLY1305
// 	case "none", "plain":
// 		return CipherType_NONE
// 	default:
// 		return CipherType_UNKNOWN
// 	}
// }

func hkdfSHA1(secret, salt, outKey []byte) {
	r := hkdf.New(sha1.New, secret, salt, []byte("ss-subkey"))
	common.Must2(io.ReadFull(r, outKey))
}
