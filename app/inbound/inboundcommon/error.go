package inboundcommon

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
)

func HandleError(ctx context.Context, err error) {
	log.Ctx(ctx).Debug().Err(err).Type("type", err).Send()
	if strings.Contains(err.Error(), "connection reset by peer") {
		return
	}
	if strings.Contains(err.Error(), "endpoint is closed for send") {
		return
	}
	if strings.Contains(err.Error(), "i/o timeout") {
		return
	}
}
