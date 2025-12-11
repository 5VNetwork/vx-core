package api

import (
	context "context"
	"encoding/base64"
	"os"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/sshhelper"
	"github.com/5vnetwork/vx-core/test"
)

// func TestParsePublicKeyFromAny(t *testing.T) {
// 	pubkey := []byte("")
// 	pub, err := ParsePublicKeyFromAny(pubkey)
// 	if err != nil {
// 		t.Fatalf("failed to parse public key: %v", err)
// 	}
// 	t.Logf("public key: %v", pub)
// }

func TestGetServerPublicKey(t *testing.T) {
	t.Skip()
	common.Must(test.LoadEnvVariables("../../.env"))
	req := &GetServerPublicKeyRequest{
		SshConfig: &ServerSshConfig{
			Address:          os.Getenv("SSH_TESTSERVER_ADDRESS"),
			Port:             22,
			Username:         os.Getenv("SSH_TESTSERVER_USERNAME"),
			SshKeyPath:       os.Getenv("SSH_TESTSERVER_SSH_KEY_PATH"),
			SshKeyPassphrase: os.Getenv("SSH_TESTSERVER_SSH_KEY_PASSPHRASE"),
		},
	}

	a := Api{}
	resp, err := a.GetServerPublicKey(context.Background(), req)
	common.Must(err)
	t.Logf("public key: %v", resp.PublicKey)
	b := base64.StdEncoding.EncodeToString(resp.PublicKey)
	t.Logf("base64 public key: %v", b)

	p, err := sshhelper.ParsePublicKeyFromAny(resp.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p.Type())
	t.Log(p)
}

// func TestServerStatusMon(t *testing.T) {
// 	common.Must(test.LoadEnvVariables("../../.env"))
// 	config := &ServerSshConfig{
// 		Address:          os.Getenv("SSH_TESTSERVER_ADDRESS"),
// 		Port:             22,
// 		Username:         os.Getenv("SSH_TESTSERVER_USERNAME"),
// 		SshKeyPath:       os.Getenv("SSH_TESTSERVER_SSH_KEY_PATH"),
// 		SshKeyPassphrase: os.Getenv("SSH_TESTSERVER_SSH_KEY_PASSPHRASE"),
// 	}

// 	req := &GetServerPublicKeyRequest{
// 		SshConfig: config,
// 	}
// 	a := Api{}
// 	resp, err := a.GetServerPublicKey(context.Background(), req)
// 	common.Must(err)
// 	t.Logf("public key: %v", resp.PublicKey)

// 	config.ServerPubKey = resp.PublicKey
// 	s, err := serverConfigToDialConfig(config)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sshClient, _, err := sshhelper.Dial(s)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer sshClient.Close()

// 	log.Info().Msg("ssh client connected")
// 	sch, err := status.GetStatusStream(context.Background(), sshClient,
// 		time.Second*time.Duration(5))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	for i := 0; i < 5; i++ {
// 		status, ok := <-sch
// 		if !ok {
// 			return
// 		}
// 		t.Log(status)
// 	}
// }

// func TestUpdateXrayConfig(t *testing.T) {
// 	test.LoadEnvVariables("../.env")
// 	client, err := test.GetTestSshClientRemote()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer client.Close()

// 	jsonConfig, err := os.ReadFile("config.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = UpdateXrayConfig(client, jsonConfig)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestInstallXray(t *testing.T) {
	t.Skip()
	common.Must(test.LoadEnvVariables("../../.env"))
	client, err := test.GetTestSshClientLocalPassword()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = InstallXray(client, os.Getenv("SSH_TESTSERVER_LOCAL_USERNAME"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstallHysteria(t *testing.T) {
	t.Skip()
	common.Must(test.LoadEnvVariables("../../.env"))
	client, err := test.GetTestSshClientLocalPassword()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = InstallHysteria(client, os.Getenv("SSH_TESTSERVER_LOCAL_USERNAME"))
	if err != nil {
		t.Fatal(err)
	}
}

// func TestReboot(t *testing.T) {
// 	test.LoadEnvVariables("../../.env")
// 	client, err := test.GetTestSshClientLocal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer client.Close()

// 	err = client.Reboot()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
