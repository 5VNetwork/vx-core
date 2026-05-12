package shadowsocks_test

import (
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/proxy/shadowsocks"
)

func TestAEADCipherUDP(t *testing.T) {
	rawAccount := &configs.ShadowsocksAccount{
		CipherType: configs.ShadowsocksCipherType_AES_128_GCM,
		User: &configs.UserConfig{
			Secret: uuid.New().String(),
		},
	}
	account, err := shadowsocks.NewMemoryAccount(
		user.NewUser(rawAccount.User.Id, rawAccount.User.Secret),
		shadowsocks.CipherType(rawAccount.CipherType), false, false)
	common.Must(err)

	cipher := account.Cipher

	key := make([]byte, cipher.KeySize())
	common.Must2(rand.Read(key))

	payload := make([]byte, 1024)
	common.Must2(rand.Read(payload))

	b1 := buf.New()
	common.Must2(b1.ReadFullFrom(rand.Reader, cipher.IVSize()))
	common.Must2(b1.Write(payload))
	common.Must(cipher.EncodePacket(key, b1))

	common.Must(cipher.DecodePacket(key, b1))
	if diff := cmp.Diff(b1.Bytes(), payload); diff != "" {
		t.Error(diff)
	}
}
