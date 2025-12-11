package util_test

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/app/util"
)

func TestGetMyIPv4(t *testing.T) {
	ip, err := util.GetMyIPv4()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ip)
}
