// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

// Example service program that beeps.
//
// The program demonstrates how to create Windows service and
// install / remove it on a computer. It also shows how to
// stop / start / pause / continue any service, and how to
// write to event log. It also shows how to use debug
// facilities available in debug package.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/5vnetwork/vx-core/win_service/service"
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

var svcName = "vx"

func main() {
	if len(os.Args) < 2 {
		usage("no enough command specified")
	}

	var err error
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "install":
		var exepath string
		if len(os.Args) == 3 {
			exepath = os.Args[2]
		} else {
			exepath, err = exePath()
			if err != nil {
				log.Fatalf("failed to get exe path: %v", err)
			}
		}
		err = service.InstallService(svcName, "vx service", exepath)
	case "remove":
		err = service.RemoveService(svcName)
	// case "start":
	// 	err = startService(svcName)
	// case "stop":
	// 	err = controlService(svcName, svc.Stop, svc.Stopped)
	// case "pause":
	// 	err = controlService(svcName, svc.Pause, svc.Paused)
	// case "continue":
	// 	err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
}

func exePath() (string, error) {
	prog := os.Args[0]
	abs, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	p := filepath.Join(filepath.Dir(abs), "vx_service.exe")
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	return "", err
}
