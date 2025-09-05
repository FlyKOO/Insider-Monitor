# Solana Insider Monitor

一款用于监控 Solana 钱包活动、检测余额变化并接收实时警报的工具。

## 社区

加入我们的 Discord 社区：
- 获取安装和配置帮助
- 分享反馈和建议
- 与其他用户交流
- 及时了解新功能和版本
- 讨论 Solana 开发

👉 [加入 Discord 服务器](https://discord.gg/7vY9ZBPdya)

## 功能特性

- 🔍 同时监控多个 Solana 钱包
- 💰 跟踪代币余额变化
- ⚡ 对显著变化实时报警
- 🔔 支持通过 Discord 发送通知
- 💾 钱包数据持久化存储
- 🛡️ 优雅处理网络中断

---

## ⚠️ 重要：RPC 端点设置

**最常见的问题是使用默认的公共 RPC 端点**，其速率限制严格，会导致扫描失败。请参考下列指南获取合适的 RPC 端点。

### 🚀 推荐 RPC 服务商（提供免费额度）

| 服务商 | 免费额度 | 速度 | 设置 |
|-------|---------|-----|-----|
| **Helius** | 100k 请求/天 | ⚡⚡⚡ | [获取免费账号](https://helius.dev) |
| **QuickNode** | 30M 请求/月 | ⚡⚡⚡ | [获取免费账号](https://quicknode.com) |
| **Triton** | 10M 请求/月 | ⚡⚡ | [获取免费账号](https://triton.one) |
| **GenesysGo** | 自定义限制 | ⚡⚡ | [获取账号](https://genesysgo.com) |

### ❌ 避免使用（速率受限）
```
❌ https://api.mainnet-beta.solana.com （默认端点，速率受限）
❌ https://api.devnet.solana.com （仅用于开发）
❌ https://solana-api.projectserum.com （速率受限）
```

### ✅ 设置 RPC 的步骤

1. **注册**任意一家上面的服务商（都是免费的）
2. **从控制台获取**你的 RPC URL
3. **更新你的 config.json**：
   ```json
   {
     "network_url": "https://your-custom-rpc-endpoint.com",
     ...
   }
   ```

---

## 快速开始

### 前置条件

- Go 1.23.2 或更高版本
- **一个专用的 Solana RPC 端点**（见上文 [RPC 设置](#⚠️-重要rpc-端点设置)）

### 安装

```bash
# 克隆仓库
git clone https://github.com/accursedgalaxy/insider-monitor
cd insider-monitor

# 安装依赖
go mod download
```

### 配置

1. 复制示例配置文件：
```bash
cp config.example.json config.json
```

2. **⚠️ 重要**：编辑 `config.json` 并替换 RPC 端点：
```json
{
    "network_url": "YOUR_DEDICATED_RPC_URL_HERE",
    "wallets": [
        "YOUR_WALLET_ADDRESS_1",
        "YOUR_WALLET_ADDRESS_2"
    ],
    "scan_interval": "1m",
    "alerts": {
        "minimum_balance": 1000,
        "significant_change": 0.20,
        "ignore_tokens": []
    },
    "discord": {
        "enabled": false,
        "webhook_url": "",
        "channel_id": ""
    }
}
```

3. **从[上述服务商](#🚀-推荐-rpc-服务商提供免费额度)获取你的 RPC 端点**并更新 `network_url`

### 配置选项说明

- `network_url`：**你的专用 RPC 端点 URL**（参见 RPC 设置章节）
- `wallets`：需要监控的 Solana 钱包地址数组
- `scan_interval`：扫描间隔时间（例如 "30s"、"1m"、"5m"）
- `alerts`：
  - `minimum_balance`：触发警报的最小代币余额
  - `significant_change`：触发警报的比例变化（0.20 = 20%）
  - `ignore_tokens`：需要忽略的代币地址数组
- `discord`：
  - `enabled`：是否启用 Discord 通知
  - `webhook_url`：Discord webhook URL
  - `channel_id`：频道 ID

### 运行监控器

```bash
go run cmd/monitor/main.go
```

#### 使用自定义配置文件
```bash
go run cmd/monitor/main.go -config path/to/config.json
```

### 警报级别

监控器根据配置的 `significant_change` 使用三个警报级别：
- 🔴 **Critical**：变化 ≥ 阈值的 5 倍
- 🟡 **Warning**：变化 ≥ 阈值的 2 倍
- 🟢 **Info**：变化 < 阈值的 2 倍

### 数据存储

监控器将钱包数据存储在 `./data` 目录中，以：
- 防止重启后产生误报
- 追踪历史变化
- 优雅处理网络中断

### 源码构建

```bash
make build
```

可执行文件将生成在 `bin` 目录。

## 🔧 故障排查

### 常见问题与解决方案

#### ❌ “Rate limit exceeded” / “Too Many Requests” 错误
**问题**：使用了速率限制严格的默认公共 RPC 端点
```
❌ Rate limit exceeded after 5 retries
```

**解决方案**：
1. 从[推荐的服务商](#🚀-推荐-rpc-服务商提供免费额度)获取免费 RPC 端点
2. 在 `config.json` 中更新端点：
   ```json
   {
     "network_url": "https://your-custom-rpc-endpoint.com",
     ...
   }
   ```

#### ❌ “Invalid wallet address format” 错误
**问题**：config.json 中的钱包地址格式错误
```
❌ invalid wallet address format at index 0: abc123
```

**解决方案**：确保钱包地址为 32-44 位的有效 Solana base58 字符串
```json
{
  "wallets": [
    "CvQk2xkXtiMj2JqqVx1YZkeSqQ7jyQkNqqjeNE1jPTfc"  ✅ 有效格式
  ]
}
```

#### ❌ “Configuration file not found” 错误
**问题**：config.json 不存在
```
❌ Configuration file not found: config.json
```

**解决方案**：
```bash
cp config.example.json config.json
```

#### ❌ “Connection check failed” 错误
**问题**：网络或 RPC 端点出现问题

**解决方案**：
1. 检查你的网络连接
2. 确认 RPC 端点 URL 是否正确
3. 尝试其他 RPC 服务商
4. 手动测试 RPC 端点：
   ```bash
   curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"getSlot"}' \
     YOUR_RPC_ENDPOINT_URL
   ```

### 获取帮助

如果你仍然遇到问题：
1. 在 [Discord 社区](https://discord.gg/7vY9ZBPdya) 寻求帮助
2. 检查日志以获取具体错误信息
3. 确保你使用的是最新版本的监控器
4. 参考上述故障排查步骤

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 许可证

本项目基于 MIT 许可证发布，详情见 LICENSE 文件。

