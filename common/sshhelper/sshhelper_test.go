package sshhelper_test

import (
	"fmt"
	"os"

	"github.com/5vnetwork/vx-core/common/sshhelper"
)

func GetTestSshClientUbuntu() (*sshhelper.Client, error) {
	client, _, err := sshhelper.Dial(&sshhelper.DialConfig{
		Addr:     fmt.Sprintf("%s:%d", os.Getenv("UBUNTU_ADDRESS"), 22),
		User:     os.Getenv("UBUNTU_USER"),
		Password: os.Getenv("UBUNTU_PASSWORD"),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
