package xsqlite

import (
	context "context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Db struct {
	secret     string
	clientConn *grpc.ClientConn
	client     DbServiceClient
}

func NewDb(secret string, serverPort uint16) (*Db, error) {
	d := &Db{
		secret: secret,
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	clientConn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%d", serverPort),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithPerRPCCredentials(d))
	if err != nil {
		return nil, err
	}
	d.clientConn = clientConn
	d.client = NewDbServiceClient(clientConn)
	// reduce latency for first request, do a request now
	go d.client.GetHandler(context.Background(), &GetHandlerRequest{Id: 1})
	return d, nil
}

func (d *Db) Close() error {
	return d.clientConn.Close()
}

func (a *Db) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"secret": a.secret}, nil
}

func (a *Db) RequireTransportSecurity() bool {
	return true
}

func (d *Db) GetHandler(id int) *OutboundHandler {
	response, err := d.client.GetHandler(context.Background(), &GetHandlerRequest{Id: int64(id)})
	if err != nil {
		log.Error().Err(err).Stack().Int("id", id).Msg("GetHandler")
		return nil
	}
	return response.ToOutboundHandler()
}

func (d *Db) GetAllHandlers() ([]*OutboundHandler, error) {
	response, err := d.client.GetAllHandlers(context.Background(), &GetAllHandlersRequest{})
	if err != nil {
		return nil, err
	}
	handlers := response.GetHandlers()
	result := make([]*OutboundHandler, len(handlers))
	for i, h := range handlers {
		result[i] = h.ToOutboundHandler()
	}
	return result, nil
}

func (d *Db) GetHandlersByGroup(group string) ([]*OutboundHandler, error) {
	response, err := d.client.GetHandlersByGroup(context.Background(),
		&GetHandlersByGroupRequest{Group: group})
	if err != nil {
		return nil, err
	}
	handlers := response.GetHandlers()
	result := make([]*OutboundHandler, len(handlers))
	for i, h := range handlers {
		result[i] = h.ToOutboundHandler()
	}
	return result, nil
}

func (d *Db) GetBatchedHandlers(batchSize int, offset int) ([]*OutboundHandler, error) {
	log.Debug().Int("batchSize", batchSize).Int("offset", offset).Msg("GetBatchedHandlers")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	response, err := d.client.GetBatchedHandlers(ctx,
		&GetBatchedHandlersRequest{BatchSize: uint32(batchSize), Offset: uint32(offset)})
	if err != nil {
		return nil, err
	}

	handlers := response.GetHandlers()
	result := make([]*OutboundHandler, len(handlers))
	for i, h := range handlers {
		result[i] = h.ToOutboundHandler()
	}
	return result, nil
}

func (d *Db) UpdateHandler(id int, m map[string]interface{}) error {
	uhr := &UpdateHandlerRequest{
		Id: int64(id),
	}
	for k, v := range m {
		switch k {
		case "ok":
			uhr.Ok = proto.Int32(int32(v.(int)))
		case "speed":
			uhr.Speed = proto.Float32(float32(v.(float64)))
		case "ping":
			uhr.Ping = proto.Int32(int32(v.(int)))
		case "support6":
			uhr.Support6 = proto.Int32(int32(v.(int)))
		case "speed_test_time":
			uhr.SpeedTestTime = proto.Int32(int32(v.(int64)))
		case "ping_test_time":
			uhr.PingTestTime = proto.Int32(int32(v.(int64)))
		case "support6_test_time":
			uhr.Support6TestTime = proto.Int32(int32(v.(int64)))
		}
	}

	_, err := d.client.UpdateHandler(context.Background(), uhr)
	if err != nil {
		return err
	}
	return nil
}

func (d *DbOutboundHandler) ToOutboundHandler() *OutboundHandler {
	o := &OutboundHandler{
		ID:               int(d.Id),
		Ok:               int(d.Ok),
		Speed:            float64(d.Speed),
		Ping:             int(d.Ping),
		PingTestTime:     int(d.PingTestTime),
		Support6:         int(d.Support6),
		Support6TestTime: int(d.Support6TestTime),
		SpeedTestTime:    int(d.SpeedTestTime),
		Config:           d.Config,
		Selected:         d.Selected,
	}
	if d.SubId != 0 {
		subId := int(d.SubId)
		o.SubId = &subId
	}
	return o
}
