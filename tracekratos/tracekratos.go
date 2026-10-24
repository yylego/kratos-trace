package tracekratos

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/grpc/metadata"
)

// Config trace middleware config
// 追踪中间件配置
type Config struct {
	TraceKeyName string                       // HTTP head name to get trace ID // 请求头键名
	NewTraceID   func(context.Context) string // Generate new trace ID // 生成新的 trace ID
	FormatArgs   func(req any) string         // Format request args // 格式化请求参数
	LogLevel     log.Level                    // Log config // 日志级别
	LogReply     bool                         // Log response and cost // 记录响应和耗时
}

// Option is a function to set Config options
// Config 的函数式选项
type Option func(*Config)

// WithLogLevel set log config
// 设置日志级别
func WithLogLevel(level log.Level) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithNewTraceID set custom trace ID creation function
// 设置自定义 trace ID 生成函数
func WithNewTraceID(fn func(context.Context) string) Option {
	return func(c *Config) {
		c.NewTraceID = fn
	}
}

// WithFormatArgs set custom args format function
// 设置自定义参数格式化函数
func WithFormatArgs(fn func(req any) string) Option {
	return func(c *Config) {
		c.FormatArgs = fn
	}
}

// WithLogReply enable response and cost logging
// 启用响应日志和耗时记录
func WithLogReply(enable bool) Option {
	return func(c *Config) {
		c.LogReply = enable
	}
}

// newRandomTraceID builds a trace ID from a nanosecond time prefix plus a random suffix, sortable and collision-safe
//
// newRandomTraceID 用纳秒时间前缀 + 随机后缀拼出 trace ID，可排序且防碰撞
func newRandomTraceID() string {
	// Time prefix keeps IDs sortable and shows the moment; rand.Text gives 128 random bits that keep IDs unique within the same instant
	// 时间前缀让 ID 可排序、能看出时刻；rand.Text 给出 128 位随机数，保证同一时刻内的唯一性
	return strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + rand.Text()
}

// NewConfig create config with trace name and options
// 使用追踪键名和选项创建配置
func NewConfig(keyName string, opts ...Option) *Config {
	cfg := &Config{
		TraceKeyName: keyName,
		NewTraceID: func(ctx context.Context) string {
			return newRandomTraceID()
		},
		FormatArgs: func(req any) string {
			return "((" + strings.ReplaceAll(ExtractArgs(req), `"`, `'`) + "))"
		},
		LogLevel: log.LevelInfo,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// NewTraceMiddleware create trace middleware
// 创建追踪中间件
func NewTraceMiddleware(config *Config, applog *slog.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if tsp, ok := transport.FromServerContext(ctx); ok {
				traceID := tsp.RequestHeader().Get(config.TraceKeyName)
				if traceID == "" {
					traceID = config.NewTraceID(ctx)
					tsp.RequestHeader().Set(config.TraceKeyName, traceID)
				}

				// Store trace ID in context
				// 将 trace ID 存入 context
				ctx = setTraceIDToContext(ctx, traceID)

				// Propagate trace ID to outgoing gRPC metadata for downstream services
				// When this service makes gRPC calls, downstream services receive the same trace ID
				// 将 trace ID 注入出站 gRPC metadata，传递给下游服务
				// 当本服务发起 gRPC 调用时，下游服务会收到相同的 trace ID
				ctx = injectOutgoingTraceID(ctx, config.TraceKeyName, traceID)

				kind := tsp.Kind().String()
				operation := tsp.Operation()

				startTime := time.Now()
				// cp from: https://github.com/go-kratos/kratos/blob/15dd2f638e3d53d059913ca83818f5843d67a277/middleware/logging/logging.go#L47
				// cp from: https://github.com/go-kratos/kratos/blob/15dd2f638e3d53d059913ca83818f5843d67a277/middleware/logging/logging.go#L87
				applog.Log(ctx, config.LogLevel, "trace",
					"kind", "server",
					"component", kind,
					"operation", operation,
					"args", config.FormatArgs(req),
					"trace", traceID,
					"startTime", startTime.Format(time.RFC3339Nano),
				)

				reply, err = handler(ctx, req)

				// Log response if enabled
				// 如果启用则记录响应
				if config.LogReply {
					latency := time.Since(startTime)
					code := 0
					if erk := errors.FromError(err); erk != nil {
						code = int(erk.Code)
					}
					applog.Log(ctx, config.LogLevel, "trace",
						"kind", "server",
						"component", kind,
						"operation", operation,
						"trace", traceID,
						"code", code,
						"latency", latency.String(),
					)
				}
				return reply, err
			}
			return handler(ctx, req)
		}
	}
}

// NewClientMiddleware create client-side trace middleware. It reads the trace ID
// the server middleware stored in context and writes it into the outgoing request
// header, so a downstream service reached over HTTP receives the same trace ID.
// gRPC calls already propagate through outgoing metadata (see injectOutgoingTraceID);
// this middleware covers the HTTP path and works for gRPC clients too.
//
// NewClientMiddleware 创建客户端追踪中间件：从 context 取出服务端中间件存入的 trace ID，
// 写进出站请求头，让经 HTTP 调用的下游服务收到同一个 trace ID。
// gRPC 调用已由出站 metadata 传播（见 injectOutgoingTraceID），本中间件补齐 HTTP 这条路，
// 对 gRPC 客户端同样适用。
func NewClientMiddleware(config *Config) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if traceID := GetTraceIDFromContext(ctx); traceID != "" {
				if tsp, ok := transport.FromClientContext(ctx); ok {
					tsp.RequestHeader().Set(config.TraceKeyName, traceID)
				}
			}
			return handler(ctx, req)
		}
	}
}

// traceCtxKey is the context storage name to get trace ID
// context 存储 trace ID 的类型
type traceCtxKey struct{}

// setTraceIDToContext stores trace ID into context
// 将 trace ID 存入 context
func setTraceIDToContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceCtxKey{}, traceID)
}

// GetTraceIDFromContext get trace ID from context
// 从 context 中获取 trace ID
func GetTraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceCtxKey{}); v != nil {
		if traceID, ok := v.(string); ok {
			return traceID
		}
	}
	return ""
}

// GetTraceID alias to GetTraceIDFromContext
// GetTraceIDFromContext 的简写
func GetTraceID(ctx context.Context) string {
	return GetTraceIDFromContext(ctx)
}

// LogTraceID returns a log.Extractor that adds the trace ID as a log field
// Wire it in so each log line carries the trace ID:
//
//	applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("request-trace"))))
//	applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("trace-uuid"))))
//
// LogTraceID 返回一个 log.Extractor，把 trace ID 作为日志字段注入
// 把它接进日志生成流程，让每一行日志都自动带上 trace ID：
//
//	applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("request-trace"))))
//	applog := log.NewLogger(log.NewHandler(log.WithExtractor(tracekratos.LogTraceID("trace-uuid"))))
func LogTraceID(keyName string) log.Extractor {
	return func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String(keyName, GetTraceIDFromContext(ctx))}
	}
}

// ExtractArgs returns the string of the req-param
// cp from: https://github.com/go-kratos/kratos/blob/15dd2f638e3d53d059913ca83818f5843d67a277/middleware/logging/logging.go#L102
// 提取请求参数的字符串表示
func ExtractArgs(req any) string {
	if redacter, ok := req.(logging.Redacter); ok {
		return redacter.Redact()
	}
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// injectOutgoingTraceID injects trace ID into outgoing gRPC metadata
// Enables cross-service trace propagation: downstream services receive the same trace ID
// Uses lowercase key name since gRPC metadata keys must be lowercase
// Reference: apmkratos newOutgoingContextWithTraceContext implementation
//
// injectOutgoingTraceID 将 trace ID 注入出站 gRPC metadata
// 实现跨服务追踪传播：下游服务收到相同的 trace ID
// 使用小写键名因为 gRPC metadata 键必须小写
// 参考：apmkratos 的 newOutgoingContextWithTraceContext 实现
func injectOutgoingTraceID(ctx context.Context, keyName string, traceID string) context.Context {
	key := strings.ToLower(keyName)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.Pairs(key, traceID)
	} else {
		md = md.Copy()
		md.Set(key, traceID)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
