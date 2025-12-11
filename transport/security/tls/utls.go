package tls

import (
	"crypto/tls"
	"fmt"
	"net"

	utls "github.com/refraction-networking/utls"
)

func (c *TlsConfig) GetUClient(conn net.Conn, tlsConfig *tls.Config) (net.Conn, error) {
	utlsConfig, err := uTLSConfigFromTLSConfig(tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to generate utls config from tls config %w", err)
	}

	preset, found := GetFingerprint(c.Imitate)
	if !found {
		return nil, fmt.Errorf("unable to get utls preset from name %s", c.Imitate)
	}

	utlsClientConn := utls.UClient(conn, utlsConfig, *preset)

	if c.NoSNI {
		err = utlsClientConn.RemoveSNIExtension()
		if err != nil {
			return nil, fmt.Errorf("unable to remove server name indication from utls client hello %w", err)
		}
	}

	err = utlsClientConn.BuildHandshakeState()
	if err != nil {
		return nil, fmt.Errorf("unable to build utls handshake state %w", err)
	}

	// ALPN is necessary for protocols like websocket to work. The uTLS setting may be overwritten on call into
	// BuildHandshakeState, so we need to check the original tls settings.
	if tlsConfig.NextProtos != nil {
		for n, v := range utlsClientConn.Extensions {
			if aplnExtension, ok := v.(*utls.ALPNExtension); ok {
				if c.ForceAlpn == ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY {
					aplnExtension.AlpnProtocols = tlsConfig.NextProtos
					break
				}
				if c.ForceAlpn == ForceALPN_NO_ALPN {
					utlsClientConn.Extensions = append(utlsClientConn.Extensions[:n], utlsClientConn.Extensions[n+1:]...)
					break
				}
			}
		}
	}

	err = utlsClientConn.BuildHandshakeState()
	if err != nil {
		return nil, fmt.Errorf("unable to build utls handshake state after modification, %w", err)
	}

	err = utlsClientConn.Handshake()
	if err != nil {
		return nil, fmt.Errorf("unable to finish utls handshake, %w", err)
	}
	return &UConn{utlsClientConn}, nil
}

type UConn struct {
	*utls.UConn
}

func (u UConn) GetConnectionApplicationProtocol() (string, error) {
	if err := u.Handshake(); err != nil {
		return "", err
	}
	return u.ConnectionState().NegotiatedProtocol, nil
}

func uTLSConfigFromTLSConfig(config *tls.Config) (*utls.Config, error) { // nolint: unparam
	uconfig := &utls.Config{
		Rand:               config.Rand,
		Time:               config.Time,
		RootCAs:            config.RootCAs,
		NextProtos:         config.NextProtos,
		InsecureSkipVerify: config.InsecureSkipVerify,
		ServerName:         config.ServerName,
	}
	return uconfig, nil
}

// var clientHelloIDMap = map[string]*utls.ClientHelloID{
// 	"randomized":        &utls.HelloRandomized,
// 	"randomizedalpn":    &utls.HelloRandomizedALPN,
// 	"randomizednoalpn":  &utls.HelloRandomizedNoALPN,
// 	"firefox_auto":      &utls.HelloFirefox_Auto,
// 	"firefox_55":        &utls.HelloFirefox_55,
// 	"firefox_56":        &utls.HelloFirefox_56,
// 	"firefox_63":        &utls.HelloFirefox_63,
// 	"firefox_65":        &utls.HelloFirefox_65,
// 	"firefox_99":        &utls.HelloFirefox_99,
// 	"firefox_102":       &utls.HelloFirefox_102,
// 	"firefox_105":       &utls.HelloFirefox_105,
// 	"chrome_auto":       &utls.HelloChrome_Auto,
// 	"chrome_58":         &utls.HelloChrome_58,
// 	"chrome_62":         &utls.HelloChrome_62,
// 	"chrome_70":         &utls.HelloChrome_70,
// 	"chrome_72":         &utls.HelloChrome_72,
// 	"chrome_83":         &utls.HelloChrome_83,
// 	"chrome_87":         &utls.HelloChrome_87,
// 	"chrome_96":         &utls.HelloChrome_96,
// 	"chrome_100":        &utls.HelloChrome_100,
// 	"chrome_102":        &utls.HelloChrome_102,
// 	"ios_auto":          &utls.HelloIOS_Auto,
// 	"ios_11_1":          &utls.HelloIOS_11_1,
// 	"ios_12_1":          &utls.HelloIOS_12_1,
// 	"ios_13":            &utls.HelloIOS_13,
// 	"ios_14":            &utls.HelloIOS_14,
// 	"android_11_okhttp": &utls.HelloAndroid_11_OkHttp,
// 	"edge_auto":         &utls.HelloEdge_Auto,
// 	"edge_85":           &utls.HelloEdge_85,
// 	"edge_106":          &utls.HelloEdge_106,
// 	"safari_auto":       &utls.HelloSafari_Auto,
// 	"safari_16_0":       &utls.HelloSafari_16_0,
// 	"360_auto":          &utls.Hello360_Auto,
// 	"360_7_5":           &utls.Hello360_7_5,
// 	"360_11_0":          &utls.Hello360_11_0,
// 	"qq_auto":           &utls.HelloQQ_Auto,
// 	"qq_11_1":           &utls.HelloQQ_11_1,
// 	//simplified
// 	"chrome":  &utls.HelloChrome_Auto,
// 	"firefox": &utls.HelloFirefox_Auto,
// 	"safari":  &utls.HelloSafari_Auto,
// 	"ios":     &utls.HelloIOS_Auto,
// 	"android": &utls.HelloAndroid_11_OkHttp,
// 	"edge":    &utls.HelloEdge_Auto,
// 	"360":     &utls.Hello360_Auto,
// 	"qq":      &utls.HelloQQ_Auto,
// }

// func GetFingerprint(name string) (*utls.ClientHelloID, error) {
// 	preset, ok := clientHelloIDMap[name]
// 	if !ok {
// 		return nil, errors.New("unknown preset name")
// 	}
// 	return preset, nil
// }

func GetFingerprint(name string) (fingerprint *utls.ClientHelloID, found bool) {
	if name == "" || name == "unknown" {
		return &utls.HelloChrome_Auto, true
	}
	if fingerprint = PresetFingerprints[name]; fingerprint != nil {
		return fingerprint, true
	}
	if fingerprint = ModernFingerprints[name]; fingerprint != nil {
		return fingerprint, true
	}
	if fingerprint = OtherFingerprints[name]; fingerprint != nil {
		return fingerprint, true
	}
	return nil, false
}

var PresetFingerprints = map[string]*utls.ClientHelloID{
	// Recommended preset options in GUI clients
	"chrome":           &utls.HelloChrome_Auto,
	"firefox":          &utls.HelloFirefox_Auto,
	"safari":           &utls.HelloSafari_Auto,
	"ios":              &utls.HelloIOS_Auto,
	"android":          &utls.HelloAndroid_11_OkHttp,
	"edge":             &utls.HelloEdge_Auto,
	"360":              &utls.Hello360_Auto,
	"qq":               &utls.HelloQQ_Auto,
	"random":           nil,
	"randomized":       nil,
	"randomizednoalpn": nil,
	"unsafe":           nil,
}

var ModernFingerprints = map[string]*utls.ClientHelloID{
	// One of these will be chosen as `random` at startup
	"hellofirefox_99":         &utls.HelloFirefox_99,
	"hellofirefox_102":        &utls.HelloFirefox_102,
	"hellofirefox_105":        &utls.HelloFirefox_105,
	"hellofirefox_120":        &utls.HelloFirefox_120,
	"hellochrome_83":          &utls.HelloChrome_83,
	"hellochrome_87":          &utls.HelloChrome_87,
	"hellochrome_96":          &utls.HelloChrome_96,
	"hellochrome_100":         &utls.HelloChrome_100,
	"hellochrome_102":         &utls.HelloChrome_102,
	"hellochrome_106_shuffle": &utls.HelloChrome_106_Shuffle,
	"hellochrome_120":         &utls.HelloChrome_120,
	"hellochrome_131":         &utls.HelloChrome_131,
	"helloios_13":             &utls.HelloIOS_13,
	"helloios_14":             &utls.HelloIOS_14,
	"helloedge_85":            &utls.HelloEdge_85,
	"helloedge_106":           &utls.HelloEdge_106,
	"hellosafari_16_0":        &utls.HelloSafari_16_0,
	"hello360_11_0":           &utls.Hello360_11_0,
	"helloqq_11_1":            &utls.HelloQQ_11_1,
}

var OtherFingerprints = map[string]*utls.ClientHelloID{
	// Golang, randomized, auto, and fingerprints that are too old
	"hellogolang":            &utls.HelloGolang,
	"hellorandomized":        &utls.HelloRandomized,
	"hellorandomizedalpn":    &utls.HelloRandomizedALPN,
	"hellorandomizednoalpn":  &utls.HelloRandomizedNoALPN,
	"hellofirefox_auto":      &utls.HelloFirefox_Auto,
	"hellofirefox_55":        &utls.HelloFirefox_55,
	"hellofirefox_56":        &utls.HelloFirefox_56,
	"hellofirefox_63":        &utls.HelloFirefox_63,
	"hellofirefox_65":        &utls.HelloFirefox_65,
	"hellochrome_auto":       &utls.HelloChrome_Auto,
	"hellochrome_58":         &utls.HelloChrome_58,
	"hellochrome_62":         &utls.HelloChrome_62,
	"hellochrome_70":         &utls.HelloChrome_70,
	"hellochrome_72":         &utls.HelloChrome_72,
	"helloios_auto":          &utls.HelloIOS_Auto,
	"helloios_11_1":          &utls.HelloIOS_11_1,
	"helloios_12_1":          &utls.HelloIOS_12_1,
	"helloandroid_11_okhttp": &utls.HelloAndroid_11_OkHttp,
	"helloedge_auto":         &utls.HelloEdge_Auto,
	"hellosafari_auto":       &utls.HelloSafari_Auto,
	"hello360_auto":          &utls.Hello360_Auto,
	"hello360_7_5":           &utls.Hello360_7_5,
	"helloqq_auto":           &utls.HelloQQ_Auto,

	// Chrome betas'
	"hellochrome_100_psk":              &utls.HelloChrome_100_PSK,
	"hellochrome_112_psk_shuf":         &utls.HelloChrome_112_PSK_Shuf,
	"hellochrome_114_padding_psk_shuf": &utls.HelloChrome_114_Padding_PSK_Shuf,
	"hellochrome_115_pq":               &utls.HelloChrome_115_PQ,
	"hellochrome_115_pq_psk":           &utls.HelloChrome_115_PQ_PSK,
	"hellochrome_120_pq":               &utls.HelloChrome_120_PQ,
}
