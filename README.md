# Captcha Verification Service

验证码校验中转服务

## 概述

本项目提供验证码校验中转服务，支持多种验证服务提供商（如 Cloudflare Turnstile）。核心验证逻辑以模块化方式实现，可被中转服务和客户端中间件共同复用。

## 架构设计

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Client Service                             │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Middleware Tool                             │   │
│  │  ┌─────────────────┐    ┌─────────────────────────────────────┐ │   │
│  │  │ Direct Provider │    │         gRPC Proxy Mode              │ │   │
│  │  │ (直接调用Provider)  │    │    (gRPC转发到中转服务)           │ │   │
│  │  └────────┬────────┘    └──────────────┬────────────────────┘ │   │
│  └───────────┼─────────────────────────────┼──────────────────────┘   │
└──────────────┼─────────────────────────────┼──────────────────────────┘
               │                             │
               │ Direct Call                 │ gRPC
               ▼                             ▼
┌──────────────────────────────┐   ┌──────────────────────────────────┐
│     Cloudflare/Other         │   │      Captcha Proxy Service       │
│     Provider API             │   │  ┌────────────────────────────┐  │
│                              │   │  │      gRPC Server            │  │
│                              │   │  │  Verify                    │  │
│                              │   │  └─────────────┬──────────────┘  │
│                              │   │                │                  │
│                              │   │  ┌─────────────▼──────────────┐  │
│                              │   │  │    Router                  │  │
│                              │   │  │    (根据Provider路由)       │  │
│                              │   │  └─────────────┬──────────────┘  │
│                              │   │                │                  │
│                              │   │  ┌─────────────▼──────────────┐  │
│                              │   │  │    Validator (复用)        │  │
│                              │   │  │    (调用Provider API)      │  │
│                              │   │  └───────────────────────────┘  │
└──────────────────────────────┘   └──────────────────────────────────┘
```

## 目录结构

本项目遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 规范，仅保留核心目录结构。

```
captcha/
├── README.md
├── go.mod
├── api/                               # API 定义 (Protobuf)
│   └── captcha/
│       └── v1/                        # v1 版本 API
│           ├── captcha.proto          # gRPC 服务定义
│           └── captcha.pb.go          # 生成的代码
├── cmd/                               # 可执行应用程序
│   └── proxy/                         # 中转服务入口
│       └── main.go
├── configs/                          # 配置文件
│   └── config.yaml
├── internal/                         # 私有代码（不可被外部导入）
│   ├── validator/                   # 核心验证逻辑
│   │   ├── validator.go            # 验证器接口定义
│   │   ├── errors.go              # 错误定义
│   │   └── providers/
│   │       ├── provider.go         # Provider 接口定义
│   │       ├── cloudflare/         # Cloudflare Turnstile 实现
│   │       │   ├── cloudflare.go
│   │       │   └── types.go
│   │       └── example/           # 示例 Provider 实现
│   │           ├── example.go
│   │           └── types.go
│   └── proxy/                      # 中转服务内部实现
│       ├── server.go               # gRPC 服务器
│       ├── handler.go              # gRPC 处理器
│       └── config.go               # 配置定义
└── pkg/                            # 可被外部使用的公共库
    └── middleware/                 # Middleware 工具包
        ├── middleware.go           # 中间件接口定义
        ├── client.go              # gRPC 客户端
        ├── gin/                   # Gin 框架中间件
        │   └── gin.go
        ├── echo/                  # Echo 框架中间件
        │   └── echo.go
        └── stdlib/                # 标准库 net/http 中间件
            └── stdlib.go
```

## gRPC API 设计

### 服务定义

```protobuf
// api/captcha.proto
syntax = "proto3";

package captcha;

option go_package = "captcha/api";

service CaptchaService {
    // 验证验证码 token
    rpc Verify(VerifyRequest) returns (VerifyResponse);
    
    // 获取支持的 Provider 列表
    rpc ListProviders(ListProvidersRequest) returns (ListProvidersResponse);
    
    // 健康检查
    rpc Health(HealthRequest) returns (HealthResponse);
}

message VerifyRequest {
    string provider = 1;            // 验证服务提供方（如 cloudflare）
    string token = 2;              // 用户验证完成的 token
    string remote_ip = 3;          // 用户 IP 地址（可选）
    map<string, string> extra = 4; // 额外参数（如 action, cdata 等）
}

message VerifyResponse {
    bool success = 1;
    string provider = 2;
    ValidationResult result = 3;
    Error error = 4;
}

message ValidationResult {
    bool success = 1;
    string challenge_ts = 2;       // ISO 8601 时间戳
    string hostname = 3;
    string action = 4;
    string cdata = 5;
    repeated string error_codes = 6;
    map<string, string> metadata = 7;
}

message Error {
    string code = 1;
    string message = 2;
    repeated string details = 3;
}

message ListProvidersRequest {}

message ListProvidersResponse {
    repeated string providers = 1;
}

message HealthRequest {}

message HealthResponse {
    string status = 1;
    string version = 2;
}
```

## 核心接口设计

### 1. Provider 接口

所有验证 Provider 实现此接口：

```go
// internal/validator/providers/provider.go
type Provider interface {
    // Name 返回 Provider 名称
    Name() string
    
    // Validate 验证 token
    // token: 用户验证完成的 token
    // remoteIP: 用户 IP 地址
    // extra: 额外参数（如 action, cdata 等）
    Validate(ctx context.Context, token, remoteIP string, extra map[string]string) (*ValidationResult, error)
}
```

### 2. ValidationResult 结构

```go
// internal/validator/validator.go
type ValidationResult struct {
    Success      bool              `json:"success"`
    ChallengeTS  time.Time         `json:"challenge_ts,omitempty"`
    Hostname     string           `json:"hostname,omitempty"`
    Action       string           `json:"action,omitempty"`
    CData        string           `json:"cdata,omitempty"`
    ErrorCodes   []string         `json:"error-codes,omitempty"`
    Metadata     map[string]any   `json:"metadata,omitempty"`
}
```

### 3. Validator 核心方法

```go
// internal/validator/validator.go
type Validator struct {
    providers map[string]Provider
}

func (v *Validator) RegisterProvider(provider Provider)
func (v *Validator) Validate(ctx context.Context, providerName, token, remoteIP string, extra map[string]string) (*ValidationResult, error)
```

## 中转服务 gRPC 接口

### Verify - 验证接口

**请求**:
```protobuf
message VerifyRequest {
    string provider = 1;            // 必填：验证服务提供方
    string token = 2;              // 必填：用户验证完成的 token
    string remote_ip = 3;          // 可选：用户 IP 地址
    map<string, string> extra = 4; // 可选：额外参数
}
```

**响应** (成功 - gRPC OK):
```protobuf
message VerifyResponse {
    bool success = 1;
    string provider = 2;
    ValidationResult result = 3;
}
```

**响应** (失败 - gRPC INVALID_ARGUMENT):
```protobuf
message VerifyResponse {
    bool success = 1;
    string provider = 2;
    Error error = 4;
}
```

### ListProviders - Provider 列表

```protobuf
message ListProvidersRequest {}

message ListProvidersResponse {
    repeated string providers = 1;
}
```

### Health - 健康检查

```protobuf
message HealthRequest {}

message HealthResponse {
    string status = 1;
    string version = 2;
}
```

## Middleware 工具包使用

Middleware 工具包位于 `pkg/middleware`，可被外部应用导入使用。

### 1. 直接调用 Provider 模式

适用于服务直接对接验证提供方：

```go
import "captcha/pkg/middleware/gin"

// 初始化
v := validator.New()
v.RegisterProvider(cloudflare.NewProvider("your-secret-key"))

// 创建中间件
middleware := gin.New(v, gin.Config{
    Provider: "cloudflare",
    Mode:     gin.ModeDirect,  // 直接调用 Provider
})
```

### 2. gRPC Proxy 模式（转发到中转服务）

适用于将验证请求通过 gRPC 转发到中转服务：

```go
import "captcha/pkg/middleware/gin"
import "captcha/pkg/middleware/client"

// 创建 gRPC 客户端
grpcClient, err := client.NewClient("captcha-proxy:9001",
    client.WithInsecure(),
)
if err != nil {
    panic(err)
}

// 创建中间件
middleware := gin.New(nil, gin.Config{
    Provider:   "cloudflare",
    Mode:       gin.ModeGRPC,  // gRPC 转发到中转服务
    GRPCClient: grpcClient,
})
```

### 完整示例

```go
package main

import (
    "captcha/pkg/middleware/gin"
    "captcha/pkg/middleware/client"
    "captcha/internal/validator"
    "captcha/internal/validator/providers/cloudflare"
    "github.com/gin-gonic/gin"
)

func main() {
    // 模式一：直接调用 Provider
    // 1. 初始化验证器
    v := validator.New()
    v.RegisterProvider(cloudflare.NewProvider("your-cloudflare-secret-key"))
    
    // 2. 创建中间件
    captchaMiddleware := gin.New(v, gin.Config{
        Provider: "cloudflare",
        Mode:     gin.ModeDirect,
    })
    
    // 模式二：gRPC 转发到中转服务
    // 1. 创建 gRPC 客户端
    grpcClient, _ := client.NewClient("captcha-proxy:9001")
    
    // 2. 创建中间件
    captchaMiddleware := gin.New(nil, gin.Config{
        Provider:   "cloudflare",
        Mode:       gin.ModeGRPC,
        GRPCClient: grpcClient,
    })
    
    // 3. 在路由中使用
    r := gin.Default()
    r.POST("/submit", captchaMiddleware, func(c *gin.Context) {
        // 验证通过的业务逻辑
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}
```

### gRPC 客户端独立使用

```go
package main

import (
    "context"
    "captcha/pkg/middleware/client"
    "captcha/api/captcha/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // 创建连接
    conn, err := grpc.Dial("localhost:9001", 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    // 创建客户端
    c := client.NewCaptchaServiceClient(conn)
    
    // 调用验证
    resp, err := c.Verify(context.Background(), &v1.VerifyRequest{
        Provider:  "cloudflare",
        Token:    "user-token",
        RemoteIp: "1.2.3.4",
        Extra:    map[string]string{"action": "login"},
    })
    
    if err != nil {
        panic(err)
    }
    
    if resp.Success {
        // 验证成功
    }
}
```

## Cloudflare Turnstile 验证详情

### 验证流程

1. **前端生成 Token**: 用户在网页上完成 Turnstile 验证后，前端获取 token
2. **Token 发送到后端**: 前端将 token 和 provider 信息放入请求 header
3. **后端验证**: 中转服务或客户端中间件调用 Provider API 进行验证
4. **返回结果**: 验证成功允许请求，验证失败拒绝请求

### 请求 Header 约定（HTTP 模式）

```javascript
// 前端示例
fetch('/api/submit', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-Captcha-Provider': 'cloudflare',
        'X-Captcha-Token': turnstileToken
    },
    body: JSON.stringify(data)
})
```

### 错误码参考

| Error Code | 说明 |
|------------|------|
| missing-input-secret | 未提供 Secret 参数 |
| invalid-input-secret | Secret key 无效或过期 |
| missing-input-response | 未提供 Token |
| invalid-input-response | Token 无效、格式错误或已过期 |
| bad-request | 请求格式错误 |
| timeout-or-duplicate | Token 已被验证过 |
| internal-error | 内部错误 |

## 配置说明

### 中转服务配置

配置文件位于 `configs/config.yaml`：

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 9001  # gRPC 端口

providers:
  cloudflare:
    enabled: true
    secret_key: "your-secret-key"
    timeout: 10s
  
  example:
    enabled: true
    secret_key: "example-key"
    timeout: 5s
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| CAPTCHA_PROXY_HOST | gRPC 服务监听地址 | 0.0.0.0 |
| CAPTCHA_PROXY_PORT | gRPC 服务监听端口 | 9001 |
| CAPTCHA_LOG_LEVEL | 日志级别 | info |
| CAPTCHA_CONFIG_PATH | 配置文件路径 | configs/config.yaml |

## 扩展新的 Provider

1. 在 `internal/validator/providers/` 下创建新目录
2. 实现 `Provider` 接口
3. 在中转服务启动时注册 Provider

```go
// internal/validator/providers/myprovider/myprovider.go
type MyProvider struct {
    secretKey string
    client    *http.Client
}

func NewProvider(secretKey string) *MyProvider {
    return &MyProvider{
        secretKey: secretKey,
        client:    &http.Client{Timeout: 10 * time.Second},
    }
}

func (p *MyProvider) Name() string {
    return "myprovider"
}

func (p *MyProvider) Validate(ctx context.Context, token, remoteIP string, extra map[string]string) (*ValidationResult, error) {
    // 实现验证逻辑
}
```

## 最佳实践

1. **安全**: Secret Key 必须保密，使用环境变量或密钥管理服务
2. **性能**: 设置合理的超时时间，实现重试逻辑
3. **监控**: 记录验证失败和异常情况
4. **错误处理**: 对用户返回友好错误信息，不暴露内部细节

## License

MIT
