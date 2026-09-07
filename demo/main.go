package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/styles"
)

func section(title string) {
	fmt.Printf("\n── %s ──\n", title)
}

func handleOrder(ctx context.Context) {
	log := tlog.FromContext(ctx)
	log.Info("validating order")
	log.Info("order validated", tlog.Duration("took", 3*time.Millisecond))
}

func main() {
	// ── Setup ─────────────────────────────────────────────────────────────
	// Pretty (forced pretty so ANSI shows even when piped), JSON, and plain
	// text loggers all write to stdout. File loggers keep the original demo
	// paths: ~/.logs/<appName>/<path>.
	prettyLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithLevel(level.Trace),
	)
	defer prettyLog.Close()

	jsonLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithJSON(),
		tlog.WithLevel(level.Trace),
	)
	defer jsonLog.Close()

	textLog := tlog.New(
		tlog.WithConsole(
			console.WithTTY(false),
		),
		tlog.WithLevel(level.Trace),
	)
	defer textLog.Close()

	fileLog := tlog.New(
		tlog.WithFile("tlog-dev-demo2", "demo.log"),
	)
	defer fileLog.Close()

	fileJSONLog := tlog.New(
		tlog.WithFile("tlog-dev-demo2", "demo.jsonl"),
		tlog.WithJSON(),
	)
	defer fileJSONLog.Close()

	// Simulate some work.
	start := time.Now()
	elapsed := time.Since(start)
	fmt.Printf("Simulated work duration: %s\n", elapsed)

	// Random duration for demonstration (0-1day) via crypto/rand.
	var seed [8]byte
	_, _ = rand.Read(seed[:])
	ttl := time.Duration(binary.LittleEndian.Uint64(seed[:])%86_400_000) * time.Millisecond
	fmt.Printf("Simulated duration: %s\n", ttl)

	// ── 1. Levels ─────────────────────────────────────────────────────────
	section("levels (Trace..Error)")
	prettyLog.Trace("tracing internals")
	prettyLog.Debug("starting up", tlog.String("env", "development"))
	prettyLog.Info("server started", tlog.String("host", "localhost"), tlog.Int("port", 8080))
	prettyLog.Warn("high memory usage", tlog.Float64("percent", 87.4))
	prettyLog.Error("request failed", tlog.Err(errors.New("connection refused")))
	// Note: Fatal logs then calls os.Exit(1), Panic logs then panics.
	// Both are intentionally not called in this demo. There are no global
	// Fatal/Panic helpers by design — use a Logger for those.

	// ── 2. All field types ────────────────────────────────────────────────
	section("typed fields")
	fields := []tlog.Field{
		tlog.Int("id", 123),
		tlog.Int64("request_id", 9876543210),
		tlog.String("name", "John Doe"),
		tlog.Bool("active", true),
		tlog.Float64("score", 99.5),
		tlog.Duration("duration_elapsed", elapsed),
		tlog.Duration("duration_ttl", ttl),
		tlog.Time("started_at", time.Now()),
		tlog.Err(errors.New("something failed")),
		tlog.Any("any_string", "raw-any"),
		tlog.Any("any_struct", map[string]any{
			"role": "admin",
			"age":  30,
		}),
	}
	prettyLog.Info("user created", fields...)
	// Nil error is safe — rendered as <nil>.
	prettyLog.Info("nil error is safe", tlog.Err(nil))

	// ── 3. Formatters: pretty / JSON / text / file ────────────────────────
	section("formatters (pretty / json / text)")
	jsonLog.Info("user created", fields...)
	textLog.Info("user created", fields...)
	fileLog.Info("user created", fields...)
	fileJSONLog.Info("user created", fields...)
	fmt.Println("file logs appended to ~/.logs/tlog-dev-demo2/demo.log and demo.jsonl")

	// ── 4. Level filtering ────────────────────────────────────────────────
	section("level filtering (WithLevel / SetLevel)")
	filtered := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithLevel(level.Warn),
	)
	defer filtered.Close()
	filtered.Debug("suppressed by Warn minimum (not emitted)")
	filtered.Info("suppressed by Warn minimum (not emitted)")
	filtered.Warn("emitted at Warn")
	filtered.Error("emitted at Error", tlog.Err(errors.New("disk full")))
	filtered.SetLevel(level.Trace)
	filtered.Debug("visible again after SetLevel(Trace)")

	// WithLevel child overrides the minimum without mutating the parent.
	quietChild := filtered.WithLevel(level.Error)
	quietChild.Warn("suppressed by child WithLevel(Error)")
	quietChild.Error("emitted by child at Error")
	filtered.Warn("parent still logs at Trace minimum")

	// ── 5. Child loggers ──────────────────────────────────────────────────
	section("child loggers (With)")
	base := prettyLog.With(
		tlog.String("service", "demo"),
		tlog.String("version", "1.0.0"),
	)
	req := base.With(tlog.String("request_id", "req-42"))
	req.Info("checkout started", tlog.String("method", "POST"), tlog.String("path", "/checkout"))
	req.Warn("slow query", tlog.Duration("query_time", 450*time.Millisecond))

	// ── 6. Caller + timestamp format ──────────────────────────────────────
	section("caller + timestamp format")
	callerLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithCaller(),
		tlog.WithLevel(level.Trace),
	)
	defer callerLog.Close()
	callerLog.Info("includes file:line:function", tlog.String("step", "caller"))

	stampedLog := tlog.New(
		tlog.WithConsole(console.WithTTY(false)),
		tlog.WithTimestampFormat(time.RFC3339),
		tlog.WithLevel(level.Trace),
	)
	defer stampedLog.Close()
	stampedLog.Info("custom timestamp layout", tlog.String("layout", time.RFC3339))

	// ── 7. Themes and styles ──────────────────────────────────────────────
	section("themes (BadgyHTTP) + WithStyles override")
	themeLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithTheme(styles.BadgyHTTP()),
		tlog.WithLevel(level.Trace),
	)
	defer themeLog.Close()
	themeLog.Info("request", tlog.String("method", "GET"), tlog.String("path", "/api/v1/users"), tlog.Int("status", 200))

	styledLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithTheme(styles.Dev()),
		tlog.WithStyles(func(s *tlog.Styles) {
			s.Message = s.Message.Bold(true)
		}),
		tlog.WithLevel(level.Trace),
	)
	defer styledLog.Close()
	styledLog.Info("bold message via WithStyles", tlog.String("theme", "Dev+override"))

	// ── 8. Context ────────────────────────────────────────────────────────
	section("context (NewContext / WithContext / FromContext)")
	root := prettyLog.With(tlog.String("service", "order-service"))
	ctx := tlog.NewContext(context.Background(), root)
	ctx = tlog.WithContext(ctx,
		tlog.String("order_id", "ord-111"),
		tlog.String("user_id", "alice"),
	)
	tlog.FromContext(ctx).Info("order processing started")
	handleOrder(ctx)
	tlog.FromContext(ctx).Info("order processing complete")

	// ── 9. Global logger ──────────────────────────────────────────────────
	section("global logger (SetDefault / Default)")
	tlog.SetDefault(tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithLevel(level.Debug),
	).With(
		tlog.String("app", "demo"),
	))
	tlog.Debug("via package helper", tlog.String("key", "value"))
	tlog.Info("server ready", tlog.Int("port", 9090))
	tlog.Warn("deprecated endpoint called", tlog.String("path", "/v1/legacy"))
	tlog.Error("upstream unavailable", tlog.String("host", "cache.internal"))
	tlog.Default().Info("retrieved via Default()", tlog.Bool("same_logger", true))

	// ── 10. Discard + error handler ───────────────────────────────────────
	section("discard + error handler")
	discardLog := tlog.New(tlog.WithDiscard())
	defer discardLog.Close()
	discardLog.Info("silently dropped (no output)")

	errLog := tlog.New(
		tlog.WithConsole(),
		tlog.WithPretty(),
		tlog.WithErrorHandler(func(err error) {
			fmt.Fprintf(os.Stderr, "tlog error: %v\n", err)
		}),
	)
	defer errLog.Close()
	errLog.Info("errors route to WithErrorHandler", tlog.String("handler", "stderr"))
}
