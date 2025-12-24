# vx-core

[English](README.md) | [中文](README.zh.md) 

---

一个源自 [v2ray-core](https://www.v2fly.org/) 的代理工具。它为我们的多平台代理客户端[VX](https://github.com/5vnetwork/vx)所使用

## 支持的代理协议

- **VMESS** - 具有 AEAD 加密的多功能协议
- **Shadowsocks** - 具有多种加密方法的经典代理协议
- **Trojan** - 基于 TLS 的代理协议
- **Hysteria2** - 基于 UDP 的协议，针对低延迟和高速度优化
- **VLESS** - 轻量级协议，支持 Vision/XTLS
- **SOCKS5** - 标准 SOCKS 代理协议
- **HTTP** - HTTP 代理支持
- **AnyTLS** - 减少 TLS in TLS 的特征
- **Dokodemo-door** - 用于端口转发

## 服务端快速开始

我们建议使用我们的代理客户端 [VX](https://github.com/5vnetwork/vx) 来安装和配置 vx-core，因为客户端具有图形界面和文档。

配置文件可以是 protobuf 文件，也可以是使用 ProtoJSON 格式的 json 文件。

1. 创建配置文件 `config.json`：

服务器 proto 配置可以在 protos/server/server.proto 中找到

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

2. 运行 vx：

```bash
vx run --config config.json
```

## 许可证

详情请参阅 [LICENSE](LICENSE) 文件。

## License Compliance

代码中包含在 "proxy/vless"、"tranport/security/reality"、"tranport/protocol/splithttp"、"app/util/x25519"、"api/reali_scanner"、"tranport/security/tls/ech" 文件夹中的代码修改自 [Xray-core](https://github.com/XTLS/Xray-core)。它按照与原始项目相同的许可证（Mozilla Public License 2.0）分发。

## 支持

如有问题、疑问或贡献，请在 GitHub 上提交 issue。谢谢

