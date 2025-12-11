//go:build linux

package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/base64"
	"sync"
	"time"
	"unsafe"

	"github.com/5vnetwork/vx-core/app/api"
	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/redirect"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/rs/zerolog/log"

	"google.golang.org/protobuf/proto"
)

var mutex sync.Mutex
var instance *client.Client
var onExit func(string)

//export Start
func Start(p unsafe.Pointer, len C.int) *C.char {
	mutex.Lock()
	defer mutex.Unlock()
	if instance != nil {
		return C.CString("close current instance first")
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

	// instance, err = builder.NewX(&config, builder.WithOnSelfStop(onExit))
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

func main() {}
