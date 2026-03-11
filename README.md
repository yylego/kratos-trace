[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-trace/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-trace/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-trace)](https://pkg.go.dev/github.com/yylego/kratos-trace)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-trace/main.svg)](https://coveralls.io/github/yylego/kratos-trace?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-trace.svg)](https://github.com/yylego/kratos-trace/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-trace)](https://goreportcard.com/report/github.com/yylego/kratos-trace)

# kratos-trace

Trace ID middleware with cross-service propagation using Kratos framework, providing request tracking and distributed tracing.

---

## CHINESE README

[中文说明](README.zh.md)

## Main Features

- 🔍 Trace ID Logging - Show trace ID in request logs
- 🔗 Cross-Service Propagation - Auto inject trace ID into outgoing gRPC metadata for downstream services
- 🚀 Auto Generation - Auto generate trace ID when not provided
- ⚙️ Flexible Config - Build config with custom options
- 📊 Response Logging - Response and cost logging on demand
- 🎯 Context Access - Get trace ID from context in business code

## Installation

```bash
go get github.com/yylego/kratos-trace/tracekratos
```

## Quick Start

```go
package main

import (
    "context"

    "github.com/go-kratos/kratos/v2/log"
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
    middleware := tracekratos.NewTraceMiddleware(config, log.DefaultLogger)

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

### Get Trace ID in Business Code

```go
func (s *Service) DoSomething(ctx context.Context) {
    traceID := tracekratos.GetTraceID(ctx)
    // or
    traceID := tracekratos.GetTraceIDFromContext(ctx)
}
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
func NewTraceMiddleware(config *Config, logger log.Logger) middleware.Middleware // Create trace middleware
func GetTraceID(ctx context.Context) string // Get trace ID from context (short name)
func GetTraceIDFromContext(ctx context.Context) string // Get trace ID from context
func ExtractArgs(req any) string // Extract request args to string
```

## License

MIT License - see [LICENSE](LICENSE) file

---

## GitHub Stars

[![Stargazers](https://starchart.cc/yylego/kratos-trace.svg?variant=adaptive)](https://starchart.cc/yylego/kratos-trace)
