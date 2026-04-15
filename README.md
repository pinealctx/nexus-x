# nexus-x

Nexus 生态公共 Go 库。提供配置加载、日志、Proto 工具、Adaptive Card 构建器，以及 Agent 引擎和 IM 客户端封装。

## Packages

| Package | Description |
|---------|-------------|
| `nxutil` | 纯工具函数 — 会话 ID 编码、HMAC、时间（零外部依赖） |
| `nxconfig` | 配置加载 — Source 抽象、YAML/JSON 自动检测、Cobra flags、TLS |
| `nxconfig/awssm` | AWS Secrets Manager 配置源 |
| `nxlog` | 全局结构化日志（zap 封装） |
| `nxproto` | Proto 工具 — 敏感字段脱敏、Connect RPC 拦截器、错误码重导出（nxerr） |
| `adaptivecard` | Adaptive Card 类型安全构建器（零外部依赖） |
| `agentic` | Agent 引擎 — Engine、Router、Middleware、Memory、Channel 接口、安全（SSRF/凭证） |
| `agentic/tools` | 内置 LLM 工具 — 消息收发、历史查询、群组管理、媒体操作，按层级分组（Basic/Query/Group/Media/All） |
| `agentic/llmconfig` | LLM 配置类型 — 模型注册表、配置加载 |
| `client` | Nexus IM 客户端 — 实现 agentic.Channel，封装 Webhook/WebSocket 接收、全部 Connect RPC 服务、Mini App initData 校验 |

## 依赖关系

```
nxutil          ← 零依赖（仅 stdlib）
nxconfig        ← yaml.v3, cobra
nxlog           ← zap
nxproto         ← nexus-proto, zap, connect
adaptivecard    ← 零依赖（仅 encoding/json）
agentic         ← nexus-proto, fantasy, adaptivecard
client          ← nexus-proto, connect, nxutil, agentic, websocket, redis
```

## 许可证

私有项目。
