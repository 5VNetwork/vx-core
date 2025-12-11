package clientgrpc

import "github.com/rs/zerolog/log"

func (d *ClientGrpc) EnableFakeDns() error {
	if d.Client.AllFakeDns != nil {
		log.Info().Msg("fake dns enabled")
		d.Client.SetFakeDnsEnabled(true)
	}
	return nil
}

func (d *ClientGrpc) DisableFakeDns() error {
	if d.Client.AllFakeDns != nil {
		log.Info().Msg("fake dns disabled")
		d.Client.SetFakeDnsEnabled(false)
	}
	return nil
}
