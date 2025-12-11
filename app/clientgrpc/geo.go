package clientgrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/common/geo/memloader"
	"github.com/rs/zerolog/log"
)

// TODO: more efficient
func (s *ClientGrpc) UpdateGeo(ctx context.Context, in *UpdateGeoRequest) (*UpdateGeoResponse, error) {
	log.Info().Msg("update geo")
	err := s.Client.Geo.UpdateGeo(in.Geo)
	if err != nil {
		return nil, fmt.Errorf("failed to create geo: %w", err)
	}
	return &UpdateGeoResponse{}, nil
}

func (s *ClientGrpc) AddGeoDomain(ctx context.Context, in *AddGeoDomainRequest) (*Receipt, error) {
	log.Info().Msg("add geo domain")
	g := s.Client.Geo.GetGeo()
	if g == nil {
		return nil, errors.New("geo not found")
	}
	err := g.AddDomain(in.DomainSetName, in.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to add geo domain: %w", err)
	}
	return &Receipt{}, nil
}

func (s *ClientGrpc) RemoveGeoDomain(ctx context.Context, in *RemoveGeoDomainRequest) (*Receipt, error) {
	log.Info().Msg("remove geo domain")
	g := s.Client.Geo.GetGeo()
	if g == nil {
		return nil, errors.New("geo not found")
	}
	err := g.RemoveDomain(in.DomainSetName, in.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to remove geo domain: %w", err)
	}
	return &Receipt{}, nil
}

func (s *ClientGrpc) ReplaceGeoDomains(ctx context.Context, in *ReplaceDomainSetRequest) (*Receipt, error) {
	log.Info().Msg("replace geo domains")
	m, err := geo.AtomicDomainSetToIndexMatcher(in.Set, memloader.New())
	if err != nil {
		return nil, fmt.Errorf("failed to replace geo domains: %w", err)
	}
	s.Client.Geo.AddDomainSet(in.Set.Name, &geo.IndexMatcherToDomainSet{
		IndexMatcher: m,
	})
	return &Receipt{}, nil
}

func (s *ClientGrpc) ReplaceGeoIPs(ctx context.Context, in *ReplaceIPSetRequest) (*Receipt, error) {
	log.Info().Msg("replace geo ips")
	m, err := geo.AtomicIpSetToIPMatcher(in.Set, memloader.New())
	if err != nil {
		return nil, fmt.Errorf("failed to replace geo ips: %w", err)
	}
	s.Client.Geo.AddIPSet(in.Set.Name, m)
	return &Receipt{}, nil
}
