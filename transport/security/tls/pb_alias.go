package tls

import tlspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/transport/security/tls"

type TlsConfig = tlspb.TlsConfig
type Certificate = tlspb.Certificate

const (
	ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY = tlspb.ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY
	ForceALPN_NO_ALPN                            = tlspb.ForceALPN_NO_ALPN
	ForceALPN_UTLS_PRESET                        = tlspb.ForceALPN_UTLS_PRESET
)
