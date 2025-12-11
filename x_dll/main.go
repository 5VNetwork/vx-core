//go:build windows

package main

/*
#include <stdlib.h>
typedef void (*self_stop_callback_t)(char*);
static inline void do_self_stop_callback(self_stop_callback_t cb, char* s) {
	cb(s);
}
typedef void (*status_change_callback_t)(int);
static void do_status_change_callback(status_change_callback_t cb, int status) {
	cb(status);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/5vnetwork/vx-core/app/api"
	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/redirect"
	"github.com/5vnetwork/vx-core/common/service"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"

	"encoding/base64"

	"google.golang.org/protobuf/proto"
)

var mutex sync.Mutex
var instance *client.Client
var onExit func(string)

//export Start
func Start(p unsafe.Pointer, len C.int, callback C.self_stop_callback_t) *C.char {
	mutex.Lock()
	defer mutex.Unlock()
	if instance != nil {
		return C.CString("close current instance first")
	}
	onExit = func(s string) {
		C.do_self_stop_callback(callback, C.CString(s))
	}
	bytes := C.GoBytes(p, len)
	var config configs.TmConfig
	var err error
	err = proto.Unmarshal(bytes, &config)
	if err != nil {
		return C.CString(err.Error())
	}

	if config.RedirectStdErr != "" {
		log.Info().Msgf("redirecting stderr to %s", config.RedirectStdErr)
		err := redirect.RedirectStderr(config.RedirectStdErr)
		if err != nil {
			log.Err(err).Msg("failed to redirect stderr")
		}
	}

	instance, err = buildclient.NewX(&config)
	if err != nil {
		redirect.CloseStderr()
		return C.CString(err.Error())
	}
	err = instance.Start()
	if err != nil {
		redirect.CloseStderr()
		instance = nil
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export Stop
func Stop() *C.char {
	mutex.Lock()
	defer mutex.Unlock()
	if instance == nil {
		return C.CString("")
	}
	err := instance.Close()
	if err != nil {
		return C.CString(err.Error())
	}
	instance = nil
	redirect.CloseStderr()
	return C.CString("")
}

//export FreeString
func FreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//export FreeBytes
func FreeBytes(p unsafe.Pointer) {
	C.free(p)
}

var apiLock sync.Mutex
var apiServer *api.Api

//export StartApiServer
func StartApiServer(p unsafe.Pointer, len C.int) *C.char {
	apiLock.Lock()
	defer apiLock.Unlock()
	if apiServer != nil {
		apiServer.Stop()
		apiServer = nil
	}
	bytes := C.GoBytes(p, len)
	var config api.ApiServerConfig
	err := proto.Unmarshal(bytes, &config)
	if err != nil {
		return C.CString(err.Error())
	}
	as, err := api.StartApiServer(&config)
	if err != nil {
		return C.CString(err.Error())
	}
	apiServer = as
	return C.CString("")
}

//export GenerateTls
func GenerateTls() (unsafe.Pointer, C.int, *C.char) {
	crt, err := cert.Generate(nil, cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
		cert.NotAfter(time.Now().Add(time.Hour*24*365)))
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}
	// Convert certificate to PEM
	certPEM := "-----BEGIN CERTIFICATE-----\n" +
		base64.StdEncoding.EncodeToString(crt.Certificate) +
		"\n-----END CERTIFICATE-----\n"

	// Convert private key to PEM
	keyPEM := "-----BEGIN PRIVATE KEY-----\n" +
		base64.StdEncoding.EncodeToString(crt.PrivateKey) +
		"\n-----END PRIVATE KEY-----\n"

	bytes, err := proto.Marshal(&tls.Certificate{
		Certificate: []byte(certPEM),
		Key:         []byte(keyPEM),
	})
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}
	return C.CBytes(bytes), C.int(len(bytes)), C.CString("")
}

//export StartService
func StartService(path *C.char, name *C.char) *C.char {
	goPath := C.GoString(path)
	serviceName := C.GoString(name)
	err := startService(goPath, serviceName)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export GetServiceStatus
func GetServiceStatus(name *C.char) *C.char {
	serviceName := C.GoString(name)
	status := getServiceStatus(serviceName)
	return C.CString(status)
}

//export StopService
func StopService(name *C.char) *C.char {
	serviceName := C.GoString(name)
	err := stopService(serviceName)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

func startService(path string, name string) error {
	m, err := service.ConnectRemote("", windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := service.OpenService(m, name,
		windows.SERVICE_START|windows.SERVICE_STOP|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start(path)
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for {
		status, err := s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
		if status.State == svc.Running {
			break
		} else if status.State == svc.Stopped {
			return errors.New("unable to start service")
		}
		if timeout.Before(time.Now()) {
			return fmt.Errorf(
				"timeout waiting for service to go to state=%d", svc.Running)
		}
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

func getServiceStatus(name string) string {
	mutex.Lock()
	defer mutex.Unlock()
	if instance != nil {
		return "running"
	}

	m, err := service.ConnectRemote("", windows.SC_MANAGER_QUERY_LOCK_STATUS)
	if err != nil {
		return err.Error()
	}
	defer m.Disconnect()
	s, err := service.OpenService(m, name, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return "stopped"
		}
		log.Err(err).Msg("OpenService")
		return "unknown"
	}
	defer s.Close()
	status, err := s.Query()
	if err != nil {
		log.Err(err).Msg("query service status")
		return "unknown"
	}
	switch status.State {
	case svc.Running:
		return "running"
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "starting"
	case svc.StopPending:
		return "stopping"
	default:
		return "unknown"
	}
}

func stopService(name string) error {
	c := svc.Stop
	to := svc.Stopped

	m, err := service.ConnectRemote("", windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := service.OpenService(m, name, windows.SERVICE_STOP|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			return nil
		}
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func main() {}
