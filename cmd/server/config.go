//go:build !android

package main

import (
	"errors"
	"os"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/common/buf"

	"google.golang.org/protobuf/encoding/protojson"
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
	} else {
		b, err = os.ReadFile(path)
	}

	jsonString := string(b)
	for oldTypeUrl, newTypeUrl := range create.OldTypeUrlToNewTypeUrl {
		jsonString = strings.ReplaceAll(jsonString, oldTypeUrl, newTypeUrl)
	}
	b = []byte(jsonString)

	err = protojson.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
