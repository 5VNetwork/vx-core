//go:build !android

package main

import (
	"errors"
	"os"

	configs "github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common/buf"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func GetConfig() (*configs.ServerConfig, error) {
	var path string
	if CfgFile != "" {
		path = CfgFile
	} else {
		return nil, errors.New("config file not set")
	}

	var config configs.ServerConfig
	var b []byte
	var err error

	if path == "stdin" {
		b, err = buf.ReadAllToBytes(os.Stdin)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(b, &config)
	} else {
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = protojson.Unmarshal(b, &config)
		if err != nil {
			err = proto.Unmarshal(b, &config)
		}
	}
	if err != nil {
		return nil, err
	}

	return &config, nil
}
