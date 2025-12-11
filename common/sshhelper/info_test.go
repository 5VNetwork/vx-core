package sshhelper_test

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/joho/godotenv"
)

func TestGetServerOS(t *testing.T) {
	// t.Skip()
	common.Must(godotenv.Load("../../.env"))
	client, err := GetTestSshClientUbuntu()
	common.Must(err)
	os, err := client.GetServerOS()
	common.Must(err)
	fmt.Println(os)
}
