<!-- TEMPLATE (EN) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-trace/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-trace/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-trace)](https://pkg.go.dev/github.com/yylego/kratos-trace)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-trace/main.svg)](https://coveralls.io/github/yylego/kratos-trace?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-trace.svg)](https://github.com/yylego/kratos-trace/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-trace)](https://goreportcard.com/report/github.com/yylego/kratos-trace)

<!-- TEMPLATE (EN) CLOSE: BADGES -->

# kratos-trace

Trace ID middleware with cross-service propagation using Kratos framework, providing request tracking and distributed tracing.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## Main Features

- 🔍 Trace ID Logging - Show trace ID in request logs
- 🔗 Cross-Service Propagation - Carry the trace ID to downstream services: auto over gRPC via outgoing metadata, and over HTTP with the client middleware (`NewClientMiddleware`)
- 🚀 Auto Generation - Auto generate trace ID when not provided
- ⚙️ Flexible Config - Build config with custom options
- 📊 Response Logging - Response and cost logging on demand
- 🎯 Context Access - Get trace ID from context in business code
- 📝 Log Integration - Use `LogTraceID(keyName)` to auto-include trace ID in each log line

## Installation

```bash
go get github.com/yylego/kratos-trace/tracekratos
```

## Quick Start

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
    // Create trace config with all options
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

    // Create trace middleware
    applog := log.NewLogger(log.NewHandler())
    middleware := tracekratos.NewTraceMiddleware(config, applog)

    // Use in Kratos server
    // httpSrv := http.NewServer(
    //     http.Middleware(middleware),
    // )
}
```

## Advanced Usage

### Custom Trace ID Generation

```go
tracekratos.WithNewTraceID(func(ctx context.Context) string {
    return uuid.New().String()
})
```

### Enable Response Logging

```go
tracekratos.WithLogReply(true)
```

### Custom Log Config

```go
tracekratos.WithLogLevel(log.LevelDebug)
```

### Custom Args Format

```go
tracekratos.WithFormatArgs(func(req any) string {
    return fmt.Sprintf("%+v", req)
})
```

### Auto Include Trace ID in Each Log Line

Use `LogTraceID(keyName)` to auto-include the trace ID in each log line. No need to pass trace ID in business code — each log statement within the same request context will contain the trace ID.

```go
// Wire LogTraceID into the logger so each line carries the trace ID
applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("request-trace"))))

// Now each log line within a request auto-includes trace ID:
// level=INFO msg=processing request-trace=ABC-123-XYZ-456
// level=INFO msg=done request-trace=ABC-123-XYZ-456
```

### Get Trace ID in Business Code

```go
func (s *Service) DoSomething(ctx context.Context) {
    traceID := tracekratos.GetTraceID(ctx)
    // or
    traceID := tracekratos.GetTraceIDFromContext(ctx)
}
```

### Propagate Trace ID over HTTP Calls

`NewTraceMiddleware` already propagates the trace ID over gRPC (via outgoing metadata). For an HTTP client call, add `NewClientMiddleware` so the current trace ID is written into the outgoing request header and the downstream service receives the same trace ID.

```go
client, _ := http.NewClient(
    context.Background(),
    http.WithEndpoint("http://127.0.0.1:8001"),
    http.WithMiddleware(
        tracekratos.NewClientMiddleware(tracekratos.NewConfig("X-Trace-ID")),
    ),
)
```

## Complete Examples

See [kratos-trace-demos](https://github.com/yylego/kratos-trace-demos) to view complete integration in Kratos projects:

- **[demo1kratos](https://github.com/yylego/kratos-trace-demos/tree/main/demo1kratos)** - Basic integration with HTTP and gRPC
- **[demo2kratos](https://github.com/yylego/kratos-trace-demos/tree/main/demo2kratos)** - Advanced usage with Wire DI

## API Reference

### Config

Config struct to trace middleware.

```go
type Config struct {
    TraceKeyName string                       // HTTP head name to get trace ID
    NewTraceID   func(context.Context) string // Generate new trace ID
    FormatArgs   func(req any) string         // Format request args
    LogLevel     log.Level                    // Log config
    LogReply     bool                         // Log response and cost
}
```

### Options

```go
func WithLogLevel(level log.Level) Option      // Set log config
func WithNewTraceID(fn func(context.Context) string) Option // Set custom trace ID creation function
func WithFormatArgs(fn func(req any) string) Option // Set custom args format function
func WithLogReply(enable bool) Option          // Enable response and cost logging
```

### Functions

```go
func NewConfig(keyName string, opts ...Option) *Config // Create config with trace name and options
func NewTraceMiddleware(config *Config, applog *slog.Logger) middleware.Middleware // Create server-side trace middleware
func NewClientMiddleware(config *Config) middleware.Middleware // Create client-side middleware to propagate trace ID into the outgoing request header (HTTP and gRPC)
func LogTraceID(keyName string) log.Extractor // Returns log.Extractor to auto-include trace ID in each log line
func GetTraceID(ctx context.Context) string // Get trace ID from context (short name)
func GetTraceIDFromContext(ctx context.Context) string // Get trace ID from context
func ExtractArgs(req any) string // Extract request args to string
```

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->
