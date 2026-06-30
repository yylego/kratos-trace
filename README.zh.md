[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-trace/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-trace/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-trace)](https://pkg.go.dev/github.com/yylego/kratos-trace)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-trace/main.svg)](https://coveralls.io/github/yylego/kratos-trace?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-trace.svg)](https://github.com/yylego/kratos-trace/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-trace)](https://goreportcard.com/report/github.com/yylego/kratos-trace)

# kratos-trace

Trace ID 中间件，支持跨服务传播，在 Kratos 框架中提供请求追踪和分布式链路追踪。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 主要功能

- 🔍 Trace ID 日志 - 在请求日志中显示 trace ID
- 🔗 跨服务传播 - 把 trace ID 传给下游服务：gRPC 经出站 metadata 自动传播，HTTP 用客户端中间件（`NewClientMiddleware`）传播
- 🚀 自动生成 - 未提供时自动生成 trace ID
- ⚙️ 灵活配置 - 使用自定义选项构建配置
- 📊 响应日志 - 可选的响应和耗时日志
- 🎯 Context 访问 - 在业务代码中从 context 获取 trace ID
- 📝 日志集成 - 使用 `LogTraceID(keyName)`，让每行日志自动包含 trace ID

## 安装

```bash
go get github.com/yylego/kratos-trace/tracekratos
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/go-kratos/kratos/v3/log"
    "github.com/google/uuid"
    "github.com/yylego/kratos-trace/tracekratos"
)

func main() {
    // 创建追踪配置，使用所有选项
    config := tracekratos.NewConfig("X-Trace-ID",
        tracekratos.WithLogLevel(log.LevelDebug),
        tracekratos.WithLogReply(true),
        tracekratos.WithNewTraceID(func(ctx context.Context) string {
            return uuid.New().String()
        }),
        tracekratos.WithFormatArgs(func(req any) string {
            return fmt.Sprintf("%+v", req)
        }),
    )

    // 创建追踪中间件
    applog := log.NewLogger(log.NewHandler())
    middleware := tracekratos.NewTraceMiddleware(config, applog)

    // 在 Kratos 服务中使用
    // httpSrv := http.NewServer(
    //     http.Middleware(middleware),
    // )
}
```

## 高级用法

### 自定义 Trace ID 生成

```go
tracekratos.WithNewTraceID(func(ctx context.Context) string {
    return uuid.New().String()
})
```

### 启用响应日志

```go
tracekratos.WithLogReply(true)
```

### 自定义日志配置

```go
tracekratos.WithLogLevel(log.LevelDebug)
```

### 自定义参数格式化

```go
tracekratos.WithFormatArgs(func(req any) string {
    return fmt.Sprintf("%+v", req)
})
```

### 让每行日志自动包含 Trace ID

使用 `LogTraceID(keyName)`，无需在业务代码中手动传递 trace ID，同一请求内的每条日志都会自动包含 trace ID。

```go
// 把 LogTraceID 接到 logger 上，让每行日志自动带上 trace ID
applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("request-trace"))))

// 之后同一请求内的每行日志都自动带上 trace ID：
// level=INFO msg=processing request-trace=ABC-123-XYZ-456
// level=INFO msg=done request-trace=ABC-123-XYZ-456
```

### 在业务代码中获取 Trace ID

```go
func (s *Service) DoSomething(ctx context.Context) {
    traceID := tracekratos.GetTraceID(ctx)
    // 或者
    traceID := tracekratos.GetTraceIDFromContext(ctx)
}
```

### 在 HTTP 调用中传播 Trace ID

`NewTraceMiddleware` 已经能经 gRPC（出站 metadata）传播 trace ID。发起 HTTP 调用时，给客户端加上 `NewClientMiddleware`，它会把当前 trace ID 写进出站请求头，让下游服务收到同一个 trace ID。

```go
client, _ := http.NewClient(
    context.Background(),
    http.WithEndpoint("http://127.0.0.1:8001"),
    http.WithMiddleware(
        tracekratos.NewClientMiddleware(tracekratos.NewConfig("X-Trace-ID")),
    ),
)
```

## 完整示例

查看 [kratos-trace-demos](https://github.com/yylego/kratos-trace-demos) 查看在实际 Kratos 项目中的完整集成：

- **[demo1kratos](https://github.com/yylego/kratos-trace-demos/tree/main/demo1kratos)** - HTTP 和 gRPC 基础集成
- **[demo2kratos](https://github.com/yylego/kratos-trace-demos/tree/main/demo2kratos)** - 使用 Wire DI 的高级用法

## API 参考

### Config

追踪中间件的配置结构体。

```go
type Config struct {
    TraceKeyName string                       // 获取 trace ID 的 HTTP 头名称
    NewTraceID   func(context.Context) string // 生成新的 trace ID
    FormatArgs   func(req any) string         // 格式化请求参数
    LogLevel     log.Level                    // 日志配置
    LogReply     bool                         // 记录响应和耗时
}
```

### 选项函数

```go
func WithLogLevel(level log.Level) Option      // 设置日志配置
func WithNewTraceID(fn func(context.Context) string) Option // 设置自定义 trace ID 生成函数
func WithFormatArgs(fn func(req any) string) Option // 设置自定义参数格式化函数
func WithLogReply(enable bool) Option          // 启用响应和耗时日志
```

### 函数

```go
func NewConfig(keyName string, opts ...Option) *Config // 使用追踪键名和选项创建配置
func NewTraceMiddleware(config *Config, applog *slog.Logger) middleware.Middleware // 创建服务端追踪中间件
func NewClientMiddleware(config *Config) middleware.Middleware // 创建客户端中间件，把 trace ID 注入出站请求头（HTTP 和 gRPC）
func LogTraceID(keyName string) log.Extractor // 返回 log.Extractor，让每行日志自动包含 trace ID
func GetTraceID(ctx context.Context) string // 从 context 获取 trace ID（简写）
func GetTraceIDFromContext(ctx context.Context) string // 从 context 获取 trace ID
func ExtractArgs(req any) string // 提取请求参数的字符串表示
```

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->

## 📄 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/yylego/kratos-trace.svg?variant=adaptive)](https://starchart.cc/yylego/kratos-trace)
