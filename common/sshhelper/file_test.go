package sshhelper_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/test"
)

func TestFileExisted(t *testing.T) {
	t.Skip()
	test.LoadEnvVariables("../.env")
	c, err := test.GetTestSshClientRemote()
	common.Must(err)
	defer c.Close()

	found, err := c.FileExisted("/tmp/test")
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		t.Fatal("file not found")
	}
}

func TestCopyFileToRemote(t *testing.T) {
	t.Skip()

	test.LoadEnvVariables("../../.env")
	c, err := GetTestSshClientUbuntu()
	common.Must(err)
	defer c.Close()

	err = c.CopyContentToRemote(bytes.NewReader([]byte("hello")), "hysteria/a", 644)
	if err != nil {
		t.Fatal(err)
	}

	o, err := c.CombinedOutput("ls -l ~/hysteria", true)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(o)
}

func TestCopyFileToRemoteSftp(t *testing.T) {
	t.Skip()

	test.LoadEnvVariables("../../.env")
	c, err := GetTestSshClientUbuntu()
	common.Must(err)
	defer c.Close()

	err = c.CopyFileToRemoteSftp("../../assets/private_geosite.dat", "./x/assets/geosite.dat", 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopyFileToRemoteSftpRemoteHost(t *testing.T) {
	t.Skip()
	test.LoadEnvVariables("../../.env")
	c, err := test.GetTestSshClientRemote()
	common.Must(err)
	defer c.Close()

	err = c.CopyFileToRemoteSftp("../../assets/private_geosite.dat", "./x/assets/geosite.dat", 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloadRemoteFileToLocal(t *testing.T) {
	t.Skip()

	test.LoadEnvVariables("../../.env")
	c, err := GetTestSshClientUbuntu()
	common.Must(err)
	defer c.Close()

	err = c.DownloadRemoteFileToLocal("a", "a")
	if err != nil {
		t.Fatal(err)
	}

}

func TestDownloadRemoteFileToLocalFolder(t *testing.T) {
	t.Skip()
	test.LoadEnvVariables("../../.env")
	c, err := GetTestSshClientUbuntu()
	common.Must(err)
	defer c.Close()

	err = c.DownloadRemoteFileToLocal("/usr/local/etc/vx/config.json", "vx_config.json")
	if err != nil {
		t.Fatal(err)
	}
}

// func TestSudo(t *testing.T) {
// 	server := test.GetTestServer()
// 	client, err := server.SSHClient()
// 	common.Must(err)
// 	defer client.Close()

// 	session, err := client.NewSession()
// 	common.Must(err)
// 	defer session.Close()

// 	// stdin, _ := session.StdinPipe()

// 	stdout, err := session.StdoutPipe()
// 	common.Must(err)
// 	stderr, err := session.StderrPipe()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	go func() {
// 		scanner := bufio.NewScanner(stdout)
// 		for {
// 			if tkn := scanner.Scan(); tkn {
// 				rcv := scanner.Bytes()

// 				raw := make([]byte, len(rcv))
// 				copy(raw, rcv)

// 				fmt.Println(string(raw))
// 			} else {
// 				if scanner.Err() != nil {
// 					fmt.Println(scanner.Err())
// 				} else {
// 					fmt.Println("io.EOF")
// 				}
// 				return
// 			}
// 		}
// 	}()

// 	go func() {
// 		scanner := bufio.NewScanner(stderr)

// 		for scanner.Scan() {
// 			fmt.Println(scanner.Text())
// 		}
// 	}()

// 	err = session.Run("echo 'lzxlnac' | sudo -S ls ")
// 	common.Must(err)

// 	// if err := session.Wait(); err != nil {
// 	// 	t.Fatalf("Failed to wait for session: %s", err)
// 	// }
// }
