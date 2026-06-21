# go-code-relay

通过 WebSocket 中继使用远程 Claude 订阅的 CLI 工具。

## 架构

```
┌─────────────────────┐      WebSocket      ┌─────────────────────┐
│  A电脑 (CLI 客户端)  │ ←──────────────→  │  B电脑 (中继服务器)   │
│                      │                    │                     │
│  ./go-code-relay     │                    │  python3 server.py  │
│  - 终端交互          │                    │  - Claude API 调用   │
│  - 彩色输出          │                    │  - 多会话管理        │
└─────────────────────┘                    └─────────────────────┘
                                                       ↑
                                              使用 B 电脑的订阅 Key
```

## 快速开始

### B 电脑：启动中继服务器

```bash
cd server
export ANTHROPIC_API_KEY=sk-ant-...
export ANTHROPIC_BASE_URL=https://api.minimaxi.com/anthropic
export ANTHROPIC_MODEL=MiniMax-M2.7
python3 server.py
```

### A 电脑：运行 CLI

```bash
# 下载或编译客户端
./go-code-relay ws://B电脑IP:8765
```

或设置环境变量：

```bash
export RELAY_SERVER_URL=ws://B电脑IP:8765
./go-code-relay
```

## 命令

- 直接输入内容与 Claude 对话
- `/clear` - 清空对话历史
- `/quit` - 退出程序

## 构建

```bash
go build -o go-code-relay ./cmd/go-code-relay/
```

## 配置

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `ANTHROPIC_API_KEY` | Claude API Key | 必须设置 |
| `ANTHROPIC_BASE_URL` | API 地址 | `https://api.minimaxi.com/anthropic` |
| `ANTHROPIC_MODEL` | 模型名称 | `MiniMax-M2.7` |
| `RELAY_HOST` | 服务器监听地址 | `0.0.0.0` |
| `RELAY_PORT` | 服务器监听端口 | `8765` |

## 注意

- API Key 通过环境变量读取，**不要硬编码**
- 服务器支持多客户端并发连接，每个客户端有独立对话历史
