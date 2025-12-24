# vx-core

[English](README.md) | [中文](README.zh.md) 

---

A proxy tool originated from [v2ray-core](https://www.v2fly.org/). It powers our
multi-platform proxy client app: [VX](https://github.com/5vnetwork/vx)

## Supported Proxy Protocols

- **VMESS** - Versatile protocol with AEAD encryption
- **Shadowsocks** - Classic proxy protocol with multiple encryption methods
- **Trojan** - TLS-based proxy protocol with Vision support
- **Hysteria2** - UDP-based protocol optimized for low-latency and high-speed
- **VLESS** - Lightweight protocol with Vision/XTLS support
- **SOCKS5** - Standard SOCKS proxy protocol
- **HTTP** - HTTP proxy support
- **AnyTLS** - Reduce characteristics of TLS in TLS
- **Dokodemo-door** - For port forwarding

## Server-Side Quick Start

We recommend to using our proxy client [VX](https://github.com/5vnetwork/vx) to
install and configure vx-core, because the client has GUI interface and
documentation.

Config file can be either protobuf file, or json file using ProtoJSON format.

1. Create a configuration file `config.json`:

The server proto config can be found at protos/server/server.proto

```json
{
    "inbounds": [
        {
            "tag": "vmess",
            "ports": [10000],
            "protocol": {
                "@type": "type.googleapis.com/x.proxy.VmessServerConfig",
                "secureEncryptionOnly": true
            }
        }
    ],
    "router": {
        "rules": [
            {
                "matchAll": true,
                "outboundTag": "direct"
            }
        ]
    },
    "outbounds": [
        {
            "tag": "direct",
            "protocol": {
                "@type": "type.googleapis.com/x.proxy.FreedomConfig"
            }
        }
    ]
}
```

2. Run vx:

```bash
vx run --config config.json
```

## License

See [LICENSE](LICENSE) file for details.

## License Compliance

The code contains in "proxy/vless", "tranport/security/reality",
"tranport/protocol/splithttp", "app/util/x25519", "api/reali_scanner",
"tranport/security/tls/ech" folder is modified from
[Xray-core](https://github.com/XTLS/Xray-core). It is distributed under the same
licence(Mozilla Public License 2.0) as the original project.

## Support

For issues, questions, or contributions, please open an issue on GitHub. Thanks
