package status

import (
	"context"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/test"
	"github.com/joho/godotenv"
)

func TestStatusStream(t *testing.T) {
	common.Must(godotenv.Load("../../../.env"))

	s, err := test.GetTestSshClientLocalPassword()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	stream, err := GetStatusStream(ctx, s, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for status := range stream {
		t.Log(status)
	}
}

func TestParsePhysMemInfo(t *testing.T) {
	info, err := ParsePhysMemInfoMac("PhysMem: 15G used (8724M wired, 1453M compressor), 77M unused.")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

// func TestParseNetStatIBN(t *testing.T) {
// 	info, err := server.ParseNetstatIbn("Name       Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
// lo0        16384 <Link#1>                      27620326     0 14082828980 27620326     0 14082828980     0
// lo0        16384 127           127.0.0.1       27620326     - 14082828980 27620326     - 14082828980     -")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(info)
// }
