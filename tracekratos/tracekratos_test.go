package tracekratos

import (
	"bytes"
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("X-B3-TRACE-ID")
	t.Log(config.NewTraceID(context.Background()))

	type reqType struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	req := &reqType{}
	require.NoError(t, gofakeit.Struct(req))
	t.Log(config.FormatArgs(req))
}

func TestNewConfigWithOptions(t *testing.T) {
	config := NewConfig("X-Trace-ID",
		WithLogLevel(log.LevelDebug),
		WithNewTraceID(func(ctx context.Context) string {
			return "custom-trace-id"
		}),
	)
	require.Equal(t, log.LevelDebug, config.LogLevel)
	require.Equal(t, "custom-trace-id", config.NewTraceID(context.Background()))
}

func TestExtractArgs(t *testing.T) {
	type reqType struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	req := &reqType{}
	require.NoError(t, gofakeit.Struct(req))
	t.Log(ExtractArgs(req))
}

func TestGetTraceID(t *testing.T) {
	ctx := context.Background()
	require.Empty(t, GetTraceID(ctx))
	require.Empty(t, GetTraceIDFromContext(ctx))

	ctx = setTraceIDToContext(ctx, "TEST-TRACE-ID")
	require.Equal(t, "TEST-TRACE-ID", GetTraceID(ctx))
	require.Equal(t, "TEST-TRACE-ID", GetTraceIDFromContext(ctx))
}

func TestLogTraceID(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Build a logger whose handler extracts the trace ID from each context
	applog := log.NewLogger(log.NewHandler(log.WithWriter(&buf), log.WithExtractor(LogTraceID("request-trace"))))

	// Simulate a request context with trace ID
	ctx := setTraceIDToContext(context.Background(), "ABC-123-XYZ-456")

	// Log with context - the trace ID should appear in output
	applog.InfoContext(ctx, "hello")

	output := buf.String()
	t.Log("log output:", output)

	// Verify trace ID appears in the log output
	require.Contains(t, output, "request-trace")
	require.Contains(t, output, "ABC-123-XYZ-456")
	require.Contains(t, output, "hello")
}

func TestLogTraceID_NoTraceInContext(t *testing.T) {
	var buf bytes.Buffer
	applog := log.NewLogger(log.NewHandler(log.WithWriter(&buf), log.WithExtractor(LogTraceID("trace-uuid"))))

	// Log without trace ID in context
	ctx := context.Background()
	applog.InfoContext(ctx, "no trace")

	output := buf.String()
	t.Log("log output:", output)

	// Should still have the key but with empty value
	require.Contains(t, output, "trace-uuid")
	require.Contains(t, output, "no trace")
}

func TestWithLogReply(t *testing.T) {
	config := NewConfig("X-Trace-ID", WithLogReply(true))
	require.True(t, config.LogReply)

	config2 := NewConfig("X-Trace-ID")
	require.False(t, config2.LogReply)
}
