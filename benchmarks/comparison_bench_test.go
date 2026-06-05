// comparison_bench_test.go — reference-only comparisons against slog, zap, and zerolog.
//
// These benchmarks are intentionally not apples-to-apples: each library has
// different design goals and trade-offs. They are included purely to give
// tlog numbers context, not as a performance claim.
//
// Each suite uses the cheapest available discard/nop logger to measure the
// library's hot-path overhead, mirroring what tlog benchmarks do with
// WithDiscard().
package benchmarks

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ── slog (stdlib) ─────────────────────────────────────────────────────────────

func BenchmarkSlog_NoFields(b *testing.B) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	b.ResetTimer()
	for b.Loop() {
		log.Info("server started")
	}
}

func BenchmarkSlog_5Fields(b *testing.B) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			slog.String("method", "GET"),
			slog.String("path", "/api/v1/users"),
			slog.Int("status", 200),
			slog.Duration("latency", 12*time.Millisecond),
			slog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkSlog_5Fields_JSON(b *testing.B) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			slog.String("method", "GET"),
			slog.String("path", "/api/v1/users"),
			slog.Int("status", 200),
			slog.Duration("latency", 12*time.Millisecond),
			slog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkSlog_Filtered(b *testing.B) {
	// Level set to Warn — Info entries are dropped before handler is called.
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	b.ResetTimer()
	for b.Loop() {
		log.Info("this is filtered out", slog.String("key", "value"))
	}
}

func BenchmarkSlog_Parallel_5Fields(b *testing.B) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("parallel write", slog.String("key", "value"))
		}
	})
}

// ── zap ───────────────────────────────────────────────────────────────────────

// zapDiscardCore builds a zap core that encodes to io.Discard.
func zapDiscardCore() zapcore.Core {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	return zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
}

func BenchmarkZap_NoFields(b *testing.B) {
	log := zap.New(zapDiscardCore())
	b.ResetTimer()
	for b.Loop() {
		log.Info("server started")
	}
}

func BenchmarkZap_5Fields(b *testing.B) {
	log := zap.New(zapDiscardCore())
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			zap.String("method", "GET"),
			zap.String("path", "/api/v1/users"),
			zap.Int("status", 200),
			zap.Duration("latency", 12*time.Millisecond),
			zap.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkZap_Filtered(b *testing.B) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(io.Discard),
		zapcore.WarnLevel, // Info is filtered
	)
	log := zap.New(core)
	b.ResetTimer()
	for b.Loop() {
		log.Info("this is filtered out", zap.String("key", "value"))
	}
}

func BenchmarkZap_Parallel_5Fields(b *testing.B) {
	log := zap.New(zapDiscardCore())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("parallel write", zap.String("key", "value"))
		}
	})
}

// ── zerolog ───────────────────────────────────────────────────────────────────

func BenchmarkZerolog_NoFields(b *testing.B) {
	log := zerolog.New(io.Discard)
	b.ResetTimer()
	for b.Loop() {
		log.Info().Msg("server started")
	}
}

func BenchmarkZerolog_5Fields(b *testing.B) {
	log := zerolog.New(io.Discard)
	b.ResetTimer()
	for b.Loop() {
		log.Info().
			Str("method", "GET").
			Str("path", "/api/v1/users").
			Int("status", 200).
			Dur("latency", 12*time.Millisecond).
			Str("request_id", "abc-123").
			Msg("request completed")
	}
}

func BenchmarkZerolog_Filtered(b *testing.B) {
	log := zerolog.New(io.Discard).Level(zerolog.WarnLevel)
	b.ResetTimer()
	for b.Loop() {
		log.Info().Str("key", "value").Msg("this is filtered out")
	}
}

func BenchmarkZerolog_Parallel_5Fields(b *testing.B) {
	log := zerolog.New(io.Discard)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info().Str("key", "value").Msg("parallel write")
		}
	})
}
