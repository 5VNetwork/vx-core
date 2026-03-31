// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	"bytes"
	context "context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	vx "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/util"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/sshhelper"
	"github.com/5vnetwork/vx-core/common/sshhelper/status"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Api) GetServerPublicKey(ctx context.Context, req *GetServerPublicKeyRequest) (*GetServerPublicKeyResponse, error) {
	s, err := serverConfigToDialConfig(req.SshConfig)
	if err != nil {
		return nil, err
	}
	sshClient, hostKey, err := sshhelper.Dial(s)
	if err != nil {
		return nil, err
	}
	defer sshClient.Close()
	return &GetServerPublicKeyResponse{PublicKey: hostKey}, nil
}

func (a *Api) MonitorServer(req *MonitorServerRequest, in Api_MonitorServerServer) error {
	sshClientCache, err := a.getSshClient(in.Context(), req.SshConfig)
	if err != nil {
		return err
	}
	defer a.DecreaseClientUser(sshClientCache)

	sshClient := sshClientCache.client
	sch, err := status.GetStatusStream(in.Context(), sshClient,
		time.Second*time.Duration(req.Interval))
	if err != nil {
		return err
	}

	for {
		select {
		case <-in.Context().Done():
			return nil
		case status, ok := <-sch:
			if !ok {
				return nil
			}
			if err := in.Send(&MonitorServerResponse{
				Cpu:         status.CpuUsage,
				UsedMemory:  status.MemAll - status.MemAvail,
				TotalMemory: status.MemAll,
				UsedDisk:    status.DiskUsed,
				TotalDisk:   status.DiskAll,
				NetInSpeed:  status.NetInSpeed,
				NetOutSpeed: status.NetOutSpeed,
				NetInUsage:  status.NetInUsage,
				NetOutUsage: status.NetOutUsage,
			}); err != nil {
				return err
			}
		}
	}
}

func serverConfigToDialConfig(config *ServerSshConfig) (*sshhelper.DialConfig, error) {
	var err error
	s := sshhelper.DialConfig{}
	s.Addr = net.JoinHostPort(config.Address, mynet.Port(config.Port).String())
	s.User = config.Username
	s.Password = config.SudoPassword

	s.PrivateKeyPassphrase = config.SshKeyPassphrase

	if len(config.SshKey) > 0 {
		s.PrivateKey = config.SshKey
	} else if len(config.SshKeyPath) > 0 {
		s.PrivateKey, err = os.ReadFile(config.SshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
	}
	if config.ServerPubKey != nil {
		s.HostKey = config.ServerPubKey
	}
	return &s, nil
}

type SshClientCacheItem struct {
	client *sshhelper.Client

	Users int
	timer *time.Timer
	key   string

	done *done.Instance
	err  error
}

func (a *Api) DecreaseClientUser(c *SshClientCacheItem) error {
	a.sshClientCacheLock.Lock()
	defer a.sshClientCacheLock.Unlock()
	c.Users--
	if c.Users == 0 {
		c.timer = time.AfterFunc(time.Second*10, func() {
			a.sshClientCacheLock.Lock()
			defer a.sshClientCacheLock.Unlock()
			if c.Users == 0 {
				log.Info().Msg("ssh client closed")
				c.client.Close()
				delete(a.sshClientCache, c.key)
			}
		})
	}
	return nil
}

func (a *Api) getSshClient(ctx context.Context, config *ServerSshConfig) (*SshClientCacheItem, error) {
	key := net.JoinHostPort(config.Address,
		mynet.Port(config.Port).String())

	// get existing
	a.sshClientCacheLock.Lock()
	cacheItem, ok := a.sshClientCache[key]
	if !ok {
		cacheItem = &SshClientCacheItem{
			key:  key,
			done: done.New(),
		}
		a.sshClientCache[key] = cacheItem
		go func() {
			defer cacheItem.done.Close()
			// create new ssh client if no existing one
			s, err := serverConfigToDialConfig(config)
			if err != nil {
				cacheItem.err = err
			} else if s.HostKey == nil {
				cacheItem.err = errors.New("no host key")
			} else {
				sshClient, _, err := sshhelper.Dial(s)
				if err != nil {
					cacheItem.err = err
				} else {
					log.Info().Msg("ssh client connected")
					cacheItem.client = sshClient
					// remove the ssh connection when the connection is closed
					go func(sshClient *sshhelper.Client) {
						sshClient.Client.Wait()
						a.sshClientCacheLock.Lock()
						delete(a.sshClientCache, key)
						a.sshClientCacheLock.Unlock()
						sshClient.Close()
					}(cacheItem.client)
					return
				}
			}
			a.sshClientCacheLock.Lock()
			delete(a.sshClientCache, key)
			a.sshClientCacheLock.Unlock()
		}()
	} else {
		if cacheItem.done.Done() {
			defer a.sshClientCacheLock.Unlock()
			if cacheItem.err != nil {
				return nil, cacheItem.err
			}
			cacheItem.Users++
			if cacheItem.timer != nil {
				cacheItem.timer.Stop()
				cacheItem.timer = nil
			}
			return cacheItem, nil
		}
	}
	a.sshClientCacheLock.Unlock()

	select {
	case <-cacheItem.done.Wait():
		if cacheItem.err != nil {
			return nil, cacheItem.err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	a.sshClientCacheLock.Lock()
	cacheItem.Users++
	a.sshClientCacheLock.Unlock()

	return cacheItem, nil

}

func (a *Api) Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)

	sshClient := sshClientCache.client
	var bbrError string
	err = sshClient.EnableBbr()
	if err != nil {
		bbrError = fmt.Errorf("failed to enable bbr: %w", err).Error()
	}

	var firewallError string
	if req.DisableFirewall {
		// disable host level firewall
		_, err = sshClient.DetectFirewall()
		if err == nil {
			err = sshClient.DisableFirewall()
			if err != nil {
				firewallError = fmt.Errorf("failed to disable firewall: %w", err).Error()
			}
		}

	}

	for name, content := range req.Files {
		err = sshClient.CopyContentToRemote(bytes.NewReader(content), name, 644)
		if err != nil {
			return nil, fmt.Errorf("failed to copy %s to remote: %w", name, err)
		}
	}

	if req.XrayConfig != nil {
		err = InstallXray(sshClient, req.SshConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to install xray: %w", err)
		}
		err = UpdateXrayConfig(sshClient, req.XrayConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update xray config: %w", err)
		}
	} else if req.HysteriaConfig != nil {
		err = InstallHysteria(sshClient, req.SshConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to install hysteria: %w", err)
		}
		err = UpdateHysteriaConfig(sshClient, req.HysteriaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update hysteria config: %w", err)
		}
	} else if req.VxConfig != nil {
		err = InstallVX(sshClient, req.SshConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to install vx: %w", err)
		}
		err = updateVXConfig(sshClient, req.VxConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to update vx config: %w", err)
		}
	}

	return &DeployResponse{
		BbrError:      bbrError,
		FirewallError: firewallError,
	}, nil
}

func (a *Api) ServerAction(ctx context.Context, req *ServerActionRequest) (*ServerActionResponse, error) {
	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)

	sshClient := sshClientCache.client
	switch req.Action {
	case ServerActionRequest_ACTION_SHUTDOWN:
		err = sshClient.Shutdown()
		if err != nil {
			return nil, err
		}
	case ServerActionRequest_ACTION_RESTART:
		err = sshClient.Reboot()
		if err != nil {
			return nil, err
		}
	}
	return &ServerActionResponse{}, nil
}

func (a *Api) VproxyStatus(ctx context.Context, req *VproxyStatusRequest) (*VproxyStatusResponse, error) {
	response := &VproxyStatusResponse{}

	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)

	sshClient := sshClientCache.client
	// vproxy installed
	installed, err := sshClient.FileExisted("/usr/local/bin/vx")
	if err != nil {
		return nil, err
	}
	if !installed {
		response.Installed = false
		return response, nil
	}
	installed, err = sshClient.FileExisted("/etc/systemd/system/vx.service")
	if err != nil {
		return nil, err
	}
	if !installed {
		response.Installed = false
		return response, nil
	}
	response.Installed = true

	// vproxy version
	result, err := sshClient.Output("vx --version", false)
	if err != nil {
		return nil, err
	}
	// vproxy version 1.1.1
	parts := strings.Split(strings.TrimRight(result, "\n"), " ")
	response.Version = parts[len(parts)-1]

	// service status
	status, err := sshClient.ServiceStatus("vx")
	if err != nil {
		return nil, err
	}
	response.StartTime = status.StartAt
	response.Memory = status.Memory
	return response, nil
}

func (a *Api) VX(ctx context.Context, req *VXRequest) (*Receipt, error) {
	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)

	if req.Install {
		err = InstallVX(sshClientCache.client, req.SshConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to install vx: %w", err)
		}
	} else if req.Uninstall {
		err = UninstallVX(sshClientCache.client)
		if err != nil {
			return nil, fmt.Errorf("failed to uninstall vx: %w", err)
		}
	} else if req.Start {
		err = sshClientCache.client.Run("systemctl start vx", true)
		if err != nil {
			return nil, fmt.Errorf("failed to start vx: %w", err)
		}
	} else if req.Stop {
		err = sshClientCache.client.Run("systemctl stop vx", true)
		if err != nil {
			return nil, fmt.Errorf("failed to stop vx: %w", err)
		}
	} else if req.Restart {
		err = sshClientCache.client.Run("systemctl restart vx", true)
		if err != nil {
			return nil, fmt.Errorf("failed to restart vx: %w", err)
		}
	} else if req.Update {
		err = UpdateVX(sshClientCache.client, req.SshConfig.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to update vx: %w", err)
		}
	}

	return &Receipt{}, nil
}

func (a *Api) ServerConfig(ctx context.Context, req *ServerConfigRequest) (*ServerConfigResponse, error) {
	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)

	sshClient := sshClientCache.client

	configBytes, err := sshClient.DownloadRemoteFileToMemory("/usr/local/etc/vx/config.json")
	if err != nil {
		return nil, err
	}

	jsonString := string(configBytes)
	for oldTypeUrl, newTypeUrl := range create.OldTypeUrlToNewTypeUrl {
		jsonString = strings.ReplaceAll(jsonString, oldTypeUrl, newTypeUrl)
	}

	var config vx.ServerConfig
	err = protojson.Unmarshal([]byte(jsonString), &config)
	if err != nil {
		return nil, err
	}

	return &ServerConfigResponse{Config: &config}, nil
}

const vxConfigPath = "/usr/local/etc/vx/config.json"

func (a *Api) UpdateServerConfig(ctx context.Context, req *UpdateServerConfigRequest) (*UpdateServerConfigResponse, error) {
	sshClientCache, err := a.getSshClient(ctx, req.SshConfig)
	if err != nil {
		return nil, err
	}
	defer a.DecreaseClientUser(sshClientCache)
	sshClient := sshClientCache.client

	err = updateVXConfig(sshClient, req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to update vx config: %w", err)
	}
	return &UpdateServerConfigResponse{}, nil
}

func updateVXConfig(sshClient *sshhelper.Client, config *vx.ServerConfig) error {
	jsonMarshaler := protojson.MarshalOptions{
		Indent: "  ",
	}
	configJson, err := jsonMarshaler.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = sshClient.CopyContentToRemoteSudo(bytes.NewReader(configJson), "/usr/local/etc/vx/config.json", 644)
	if err != nil {
		return fmt.Errorf("failed to copy config to remote: %w", err)
	}

	err = sshClient.Run("systemctl restart vx", true)
	if err != nil {
		return fmt.Errorf("failed to restart vx: %w", err)
	}
	return nil
}

func (a *Api) InboundConfigToOutboundConfig(ctx context.Context,
	req *InboundConfigToOutboundConfigRequest) (*InboundConfigToOutboundConfigResponse, error) {
	if req.Inbound != nil {
		outboundConfigs, err := util.InboundConfigToOutboundConfig(req.ServerName,
			req.Inbound, req.ServerAddress)
		if err != nil {
			return nil, err
		}
		return &InboundConfigToOutboundConfigResponse{OutboundConfigs: outboundConfigs}, nil
	} else {
		outboundConfigs, err := util.MultiInboundConfigToOutboundConfig(req.ServerName,
			req.MultiInbound, req.ServerAddress)
		if err != nil {
			return nil, err
		}
		return &InboundConfigToOutboundConfigResponse{OutboundConfigs: outboundConfigs}, nil
	}
}
