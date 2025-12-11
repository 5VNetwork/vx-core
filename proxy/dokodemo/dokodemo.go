package dokodemo

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type Door struct {
	DoorSettings
}

type DoorSettings struct {
	Address  net.Address
	Port     net.Port
	Networks []net.Network
	Handler  i.Handler
}

// Init initializes the Door instance with necessary parameters.
func New(settings DoorSettings) *Door {
	d := &Door{
		DoorSettings: settings,
	}
	return d
}

func (d *Door) Network() []net.Network {
	return d.Networks
}

func (d *Door) Process(ctx context.Context, conn net.Conn) error {
	log.Ctx(ctx).Debug().Msg("dokodemo process")

	dest := net.Destination{
		Network: net.DestinationFromAddr(conn.LocalAddr()).Network,
		Address: d.Address,
		Port:    d.Port,
	}

	if !dest.IsValid() || dest.Address == nil {
		return errors.New("unable to get destination")
	}

	var reader buf.Reader
	if dest.Network == net.Network_UDP {
		reader = buf.NewPacketReader(conn)
	} else {
		reader = buf.NewReader(conn)
	}

	var writer buf.Writer
	if dest.Network == net.Network_TCP {
		writer = buf.NewWriter(conn)
	} else {
		writer = &buf.SequentialWriter{Writer: conn}
	}
	if err := d.Handler.HandleFlow(ctx, dest, buf.NewRWD(reader, writer, conn)); err != nil {
		return fmt.Errorf("failed to dispatch reader writer, %w", err)
	}
	return nil
}
