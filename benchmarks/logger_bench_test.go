// Package benchmarks provides baseline performance measurements for tlog.
// Run with: go test -bench=. -benchmem ./benchmarks/
package benchmarks

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/formatter/pretty"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/styles"
)

// ── Baseline: Info with no fields ────────────────────────────────────────────

func BenchmarkInfo_NoFields_Text(b *testing.B) {
	log := tlog.New(tlog.WithDiscard())
	b.ResetTimer()
	for b.Loop() {
		log.Info("server started")
	}
}

func BenchmarkInfo_NoFields_JSON(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithJSON())
	b.ResetTimer()
	for b.Loop() {
		log.Info("server started")
	}
}

// ── With 5 fields (typical case) ─────────────────────────────────────────────

func BenchmarkInfo_5Fields_Text(b *testing.B) {
	log := tlog.New(tlog.WithDiscard())
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			tlog.String("method", "GET"),
			tlog.String("path", "/api/v1/users"),
			tlog.Int("status", 200),
			tlog.Duration("latency", 12*time.Millisecond),
			tlog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkInfo_5Fields_JSON(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithJSON())
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			tlog.String("method", "GET"),
			tlog.String("path", "/api/v1/users"),
			tlog.Int("status", 200),
			tlog.Duration("latency", 12*time.Millisecond),
			tlog.String("request_id", "abc-123"),
		)
	}
}

// ── With 10 fields (heavy case) ───────────────────────────────────────────────

func BenchmarkInfo_10Fields_Text(b *testing.B) {
	log := tlog.New(tlog.WithDiscard())
	b.ResetTimer()
	for b.Loop() {
		log.Info("heavy entry",
			tlog.String("a", "1"),
			tlog.String("b", "2"),
			tlog.Int("c", 3),
			tlog.Int64("d", 4),
			tlog.Float64("e", 5.0),
			tlog.Bool("f", true),
			tlog.Duration("g", time.Second),
			tlog.Time("h", time.Now()),
			tlog.Err(errors.New("oops")),
			tlog.Any("i", struct{ X int }{X: 42}),
		)
	}
}

func BenchmarkInfo_10Fields_JSON(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithJSON())
	b.ResetTimer()
	for b.Loop() {
		log.Info("heavy entry",
			tlog.String("a", "1"),
			tlog.String("b", "2"),
			tlog.Int("c", 3),
			tlog.Int64("d", 4),
			tlog.Float64("e", 5.0),
			tlog.Bool("f", true),
			tlog.Duration("g", time.Second),
			tlog.Time("h", time.Now()),
			tlog.Err(errors.New("oops")),
			tlog.Any("i", struct{ X int }{X: 42}),
		)
	}
}

// ── Level filtering ──────────────────────────────────────────────────────────

func BenchmarkDebug_Filtered(b *testing.B) {
	// Level set to INFO — Debug entries are dropped before formatting.
	log := tlog.New(tlog.WithDiscard(), tlog.WithLevel(level.Info))
	b.ResetTimer()
	for b.Loop() {
		log.Debug("this is filtered out",
			tlog.String("key", "value"),
		)
	}
}

// ── With caller info ─────────────────────────────────────────────────────────

func BenchmarkInfo_WithCaller_Text(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithCaller())
	b.ResetTimer()
	for b.Loop() {
		log.Info("caller enabled")
	}
}

func BenchmarkInfo_WithCaller_JSON(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithJSON(), tlog.WithCaller())
	b.ResetTimer()
	for b.Loop() {
		log.Info("caller enabled")
	}
}

// ── Console vs file output ────────────────────────────────────────────────────

// BenchmarkInfo_ConsoleOutput writes through the console output struct
// (mutex + formatter dispatch) to a null writer — isolates output overhead
// from I/O cost.
func BenchmarkInfo_ConsoleOutput_Text(b *testing.B) {
	log := tlog.New(
		tlog.WithConsole(console.WithWriter(io.Discard), console.WithTTY(false)),
	)
	b.ResetTimer()
	for b.Loop() {
		log.Info("console write", tlog.String("key", "value"))
	}
}

func BenchmarkInfo_ConsoleOutput_JSON(b *testing.B) {
	log := tlog.New(
		tlog.WithConsole(console.WithWriter(io.Discard), console.WithTTY(false)),
		tlog.WithJSON(),
	)
	b.ResetTimer()
	for b.Loop() {
		log.Info("console write", tlog.String("key", "value"))
	}
}

func BenchmarkInfo_FileOutput_JSON(b *testing.B) {
	log := tlog.New(tlog.WithFile(b.TempDir()+"/bench.log"), tlog.WithJSON())
	b.Cleanup(func() { log.Close() })
	b.ResetTimer()
	for b.Loop() {
		log.Info("file write", tlog.String("key", "value"))
	}
}

// ── Parallel ─────────────────────────────────────────────────────────────────

func BenchmarkInfo_Parallel_Text(b *testing.B) {
	log := tlog.New(tlog.WithDiscard())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("parallel write", tlog.String("key", "value"))
		}
	})
}

func BenchmarkInfo_Parallel_JSON(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithJSON())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("parallel write", tlog.String("key", "value"))
		}
	})
}

// ── Pretty formatter (M3 regression check) ───────────────────────────────────

func BenchmarkInfo_NoFields_Pretty(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithFormatter(pretty.New()))
	b.ResetTimer()
	for b.Loop() {
		log.Info("server started")
	}
}

func BenchmarkInfo_5Fields_Pretty(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithFormatter(pretty.New()))
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			tlog.String("method", "GET"),
			tlog.String("path", "/api/v1/users"),
			tlog.Int("status", 200),
			tlog.Duration("latency", 12*time.Millisecond),
			tlog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkInfo_5Fields_Pretty_NoColor(b *testing.B) {
	log := tlog.New(
		tlog.WithDiscard(),
		tlog.WithFormatter(pretty.New(pretty.WithStyles(styles.NoColor()))),
	)
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			tlog.String("method", "GET"),
			tlog.String("path", "/api/v1/users"),
			tlog.Int("status", 200),
			tlog.Duration("latency", 12*time.Millisecond),
			tlog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkInfo_5Fields_Pretty_Plain(b *testing.B) {
	// isTTY=false — exercises the plain-text fallback path in PrettyFormatter.
	log := tlog.New(
		tlog.WithDiscard(),
		tlog.WithFormatter(pretty.New(pretty.WithTTY(false))),
	)
	b.ResetTimer()
	for b.Loop() {
		log.Info("request completed",
			tlog.String("method", "GET"),
			tlog.String("path", "/api/v1/users"),
			tlog.Int("status", 200),
			tlog.Duration("latency", 12*time.Millisecond),
			tlog.String("request_id", "abc-123"),
		)
	}
}

func BenchmarkInfo_10Fields_Pretty(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithFormatter(pretty.New()))
	b.ResetTimer()
	for b.Loop() {
		log.Info("heavy entry",
			tlog.String("a", "1"),
			tlog.String("b", "2"),
			tlog.Int("c", 3),
			tlog.Int64("d", 4),
			tlog.Float64("e", 5.0),
			tlog.Bool("f", true),
			tlog.Duration("g", time.Second),
			tlog.Time("h", time.Now()),
			tlog.Err(errors.New("oops")),
			tlog.Any("i", struct{ X int }{X: 42}),
		)
	}
}

func BenchmarkInfo_Parallel_Pretty(b *testing.B) {
	log := tlog.New(tlog.WithDiscard(), tlog.WithFormatter(pretty.New()))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("parallel write", tlog.String("key", "value"))
		}
	})
}
