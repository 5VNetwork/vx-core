package reality

import (
	"io"
	"os"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/rs/zerolog/log"
	"github.com/xtls/reality"
)

func (c *RealityConfig) GetREALITYConfig() *reality.Config {
	var dialer net.Dialer
	config := &reality.Config{
		DialContext:  dialer.DialContext,
		Type:         "tcp",
		Show:         c.Show,
		Dest:         c.Dest,
		Xver:         byte(c.Xver),
		PrivateKey:   c.PrivateKey,
		MinClientVer: c.MinClientVer,
		MaxClientVer: c.MaxClientVer,
		MaxTimeDiff:  time.Duration(c.MaxTimeDiff) * time.Millisecond,

		NextProtos:             nil, // should be nil
		SessionTicketsDisabled: true,

		KeyLogWriter: KeyLogWriterFromConfig(c),
	}
	config.ServerNames = make(map[string]bool)
	for _, serverName := range c.ServerNames {
		config.ServerNames[serverName] = true
	}
	config.ShortIds = make(map[[8]byte]bool)
	for _, shortId := range c.ShortIds {
		config.ShortIds[*(*[8]byte)(shortId)] = true
	}
	return config
}

func KeyLogWriterFromConfig(c *RealityConfig) io.Writer {
	if len(c.MasterKeyLog) <= 0 || c.MasterKeyLog == "none" {
		return nil
	}

	writer, err := os.OpenFile(c.MasterKeyLog, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Err(err).Msgf("failed to open %s as master key log", c.MasterKeyLog)
	}

	return writer
}
