package wireguard

import (
	"buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/wireguard"
	"github.com/rs/zerolog/log"
)

func createTun(c *wireguard.DeviceConfig) tunCreator {
	if !c.IsClient {
		// See tun_linux.go createKernelTun()
		log.Warn().Msg("Using gVisor TUN. WG inbound doesn't support kernel TUN yet.")
		return createGVisorTun
	}
	if c.NoKernelTun {
		log.Warn().Msg("Using gVisor TUN. NoKernelTun is set to true.")
		return createGVisorTun
	}
	kernelTunSupported, err := KernelTunSupported()
	if err != nil {
		log.Warn().Msgf("Using gVisor TUN. Failed to check kernel TUN support: %v", err)
		return createGVisorTun
	}
	if !kernelTunSupported {
		log.Warn().Msg("Using gVisor TUN. Kernel TUN is not supported on your OS, or your permission is insufficient.")
		return createGVisorTun
	}
	log.Warn().Msg("Using kernel TUN.")
	return createKernelTun
}
