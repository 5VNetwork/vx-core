package sshhelper_test

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/joho/godotenv"
)

func TestServiceStatus(t *testing.T) {
	t.Skip()
	common.Must(godotenv.Load("../../.env"))
	client, err := GetTestSshClientUbuntu()
	common.Must(err)
	status, err := client.ServiceStatus("vx")
	common.Must(err)
	fmt.Println(status)
}
