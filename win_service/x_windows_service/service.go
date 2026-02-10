// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/grpcservice"
	"github.com/5vnetwork/vx-core/common/redirect"
	"github.com/5vnetwork/vx-core/win_service/wfp"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"google.golang.org/protobuf/encoding/protojson"
)

var elog debug.Log

type exampleService struct {
	instance *client.Client
}

// args should be config path
func (m *exampleService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown /* | svc.AcceptPauseAndContinue */
	changes <- svc.Status{State: svc.StartPending}

	elog.Info(1, fmt.Sprintf("args: %v", args))

	configPath := ""
	if len(args) <= 1 {
		configPath = "config.json"
	} else {
		configPath = args[1]
	}

	var config configs.TmConfig
	b, err := os.ReadFile(configPath)
	if err != nil {
		changes <- svc.Status{State: svc.StopPending}
		elog.Error(1, fmt.Sprintf("failed to read config: %v", err))
		return
	}
	if strings.HasSuffix(configPath, ".json") {
		err = protojson.Unmarshal(b, &config)
	} else {
		err = proto.Unmarshal(b, &config)
	}
	if err != nil {
		changes <- svc.Status{State: svc.StopPending}
		elog.Error(1, fmt.Sprintf("failed to unmarshal config: %v", err))
		return
	}
	if config.RedirectStdErr != "" {
		elog.Info(1, fmt.Sprintf("redirecting stderr to %s", config.RedirectStdErr))
		err := redirect.RedirectStderr(config.RedirectStdErr)
		if err != nil {
			elog.Error(1, fmt.Sprintf("failed to redirect stderr: %v", err))
			// return
		} else {
			elog.Info(1, fmt.Sprintf("redirected stderr to %s", config.RedirectStdErr))
		}
		// defer redirect.CloseStderr()
	}

	instance, err := buildclient.NewX(&config)
	if err != nil {
		changes <- svc.Status{State: svc.StopPending}
		elog.Error(1, fmt.Sprintf("failed to create client: %v", err))
		return
	}
	m.instance = instance

	clientGrpc := m.instance.Components.GetComponent(reflect.TypeOf(&grpcservice.GrpcService{})).(*grpcservice.GrpcService)
	clientGrpc.RunningInService = true
	exitChan := make(chan struct{})
	clientGrpc.OnExit = func() {
		exitChan <- struct{}{}
	}
	instance.Components.AddComponent(wfp.New(config.GetTun().GetDevice().GetName()))

	err = instance.Start()
	if err != nil {
		changes <- svc.Status{State: svc.StopPending}
		elog.Error(1, fmt.Sprintf("failed to start client: %v", err))
		return
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case <-exitChan:

			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				testOutput := strings.Join(args, "-")
				testOutput += fmt.Sprintf("-%d", c.Context)
				elog.Info(1, testOutput)
				break loop
			// case svc.Pause:
			// 	changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			// 	tick = slowtick
			// case svc.Continue:
			// 	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			// 	tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	m.instance.Close()
	m.instance = nil
	changes <- svc.Status{State: svc.StopPending}

	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &exampleService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
