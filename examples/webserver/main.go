// examples/webserver — HTTP access log and structured middleware with tlog.
//
// Demonstrates a realistic web server setup:
//   - Per-request child logger with request-scoped fields (method, path, request_id).
//   - HTTP access log middleware that records method, path, status, latency.
//   - Panic recovery middleware that logs stack traces.
//   - Service-level context fields (service, version) on every line.
//   - Four simultaneous outputs via two loggers:
//     1. Pretty console → stderr       (human, dev view)
//     2. Plain text file → app.log     (human, ops tail -f)
//     3. JSON file → app.json          (machine, log aggregators)
//     4. JSON → stdout                 (machine, container stdout stream)
//
// Because tlog uses one formatter per logger, human (text/pretty) and machine
// (JSON) outputs are managed by two loggers combined in a dualLog helper.
//
// Both files are written to os.TempDir(). Their paths are printed on startup.
//
// Run:
//
//	go run ./examples/webserver/
package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/tidjee-dev/tlog"
	"github.com/tidjee-dev/tlog/level"
	"github.com/tidjee-dev/tlog/outputs/console"
	"github.com/tidjee-dev/tlog/outputs/file"
)

// ── dualLog ───────────────────────────────────────────────────────────────────
//
// dualLog fans every log call to two underlying tlog.Loggers:
//   - human: pretty console (stderr) + plain text file  (text/pretty formatter)
//   - mach:  JSON stdout    + JSON file                  (JSON formatter)

type dualLog struct {
	human *tlog.Logger
	mach  *tlog.Logger
}

func (d *dualLog) With(fields ...tlog.Field) *dualLog {
	return &dualLog{human: d.human.With(fields...), mach: d.mach.With(fields...)}
}
func (d *dualLog) Debug(msg string, fields ...tlog.Field) {
	d.human.Debug(msg, fields...)
	d.mach.Debug(msg, fields...)
}
func (d *dualLog) Info(msg string, fields ...tlog.Field) {
	d.human.Info(msg, fields...)
	d.mach.Info(msg, fields...)
}
func (d *dualLog) Warn(msg string, fields ...tlog.Field) {
	d.human.Warn(msg, fields...)
	d.mach.Warn(msg, fields...)
}
func (d *dualLog) Error(msg string, fields ...tlog.Field) {
	d.human.Error(msg, fields...)
	d.mach.Error(msg, fields...)
}
func (d *dualLog) Close() {
	d.human.Close()
	d.mach.Close()
}

// ── context ───────────────────────────────────────────────────────────────────

type contextKey int

const loggerKey contextKey = 0

func loggerFrom(ctx context.Context) *dualLog {
	if l, ok := ctx.Value(loggerKey).(*dualLog); ok {
		return l
	}
	noop := tlog.New()
	return &dualLog{human: noop, mach: noop}
}

// ── middleware ────────────────────────────────────────────────────────────────

// withRequestLogger injects a request-scoped child dualLog into the context.
func withRequestLogger(base *dualLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var b [4]byte
			_, _ = rand.Read(b[:])
			reqID := fmt.Sprintf("%08x", binary.LittleEndian.Uint32(b[:]))
			log := base.With(
				tlog.String("request_id", reqID),
				tlog.String("method", r.Method),
				tlog.String("path", r.URL.Path),
				tlog.String("remote_addr", r.RemoteAddr),
			)
			ctx := context.WithValue(r.Context(), loggerKey, log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// accessLog records status code and latency after each request.
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		elapsed := time.Since(start)

		log := loggerFrom(r.Context()).With(
			tlog.Int("status", rw.status),
			tlog.Duration("latency", elapsed),
		)
		switch {
		case rw.status >= 500:
			log.Error("request served")
		case rw.status >= 400:
			log.Warn("request served")
		default:
			log.Info("request served")
		}
	})
}

// panicRecovery catches panics, logs them with a stack trace, and returns 500.
func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				loggerFrom(r.Context()).Error("panic recovered",
					tlog.Any("panic", rec),
					tlog.String("stack", string(debug.Stack())),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the written status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// ── handlers ──────────────────────────────────────────────────────────────────

func handleHealth(w http.ResponseWriter, r *http.Request) {
	loggerFrom(r.Context()).Debug("health check")
	_, _ = fmt.Fprintln(w, `{"status":"ok"}`)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	log.Debug("fetching users", tlog.String("query", r.URL.RawQuery))

	// #nosec G404
	delay := time.Duration(mathrand.IntN(80)+10) * time.Millisecond
	time.Sleep(delay)
	if delay > 60*time.Millisecond {
		log.Warn("slow query detected", tlog.Duration("query_time", delay))
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintln(w, `[{"id":1,"name":"alice"},{"id":2,"name":"bob"}]`)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Warn("missing order id")
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("invalid order id", tlog.String("raw_id", idStr), tlog.Err(err))
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	log.Info("order fetched", tlog.Int("order_id", id))
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"id":%d,"status":"shipped"}`, id)
}

func handlePanic(_ http.ResponseWriter, r *http.Request) {
	loggerFrom(r.Context()).Warn("about to panic (intentional demo)")
	panic(errors.New("something went very wrong"))
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	dir := os.TempDir()
	textPath := filepath.Join(dir, "tlog-webserver.log")
	jsonPath := filepath.Join(dir, "tlog-webserver.jsonl")

	// ── human logger: pretty console (stderr) + plain text file ──────────────
	textFile, err := file.New(textPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open text log: %v\n", err)
		os.Exit(1)
	}
	humanLog := tlog.New(
		tlog.WithConsole(console.WithWriter(os.Stderr), console.WithTTY(true)),
		tlog.WithOutput(textFile),
		tlog.WithLevel(level.Debug),
	)

	// ── machine logger: JSON stdout + JSON file ───────────────────────────────
	jsonFile, err := file.New(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open json log: %v\n", err)
		os.Exit(1)
	}
	machLog := tlog.New(
		tlog.WithConsole(console.WithWriter(os.Stdout), console.WithTTY(false)),
		tlog.WithOutput(jsonFile),
		tlog.WithJSON(),
		tlog.WithLevel(level.Debug),
	)

	// ── combined logger with service-level fields ─────────────────────────────
	log := (&dualLog{human: humanLog, mach: machLog}).With(
		tlog.String("service", "example-api"),
		tlog.String("version", "1.0.0"),
	)
	defer log.Close()

	log.Info("log files opened",
		tlog.String("text_log", textPath),
		tlog.String("json_log", jsonPath),
	)

	// ── HTTP server ───────────────────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/users", handleUsers)
	mux.HandleFunc("/orders", handleOrders)
	mux.HandleFunc("/panic", handlePanic)

	chain := withRequestLogger(log)(panicRecovery(accessLog(mux)))

	lc := net.ListenConfig{}

	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		log.Error("failed to listen", tlog.Err(err))
		os.Exit(1)
	}
	addr := ln.Addr().String()
	srv := &http.Server{
		Handler:           chain,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info("server listening", tlog.String("addr", addr))

	// Fire canned requests then shut down.
	go func() {
		time.Sleep(100 * time.Millisecond)

		routes := []string{
			"/health",
			"/users",
			"/users?page=2",
			"/orders?id=42",
			"/orders",         // missing id → 400
			"/orders?id=oops", // bad id  → 400
			"/panic",          // recovered panic → 500
		}

		client := &http.Client{Timeout: 5 * time.Second}
		base := "http://" + addr
		for _, path := range routes {
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				base+path,
				nil,
			)

			resp, reqErr := client.Do(req)
			if reqErr == nil {
				_ = resp.Body.Close()
			}
			time.Sleep(20 * time.Millisecond)
		}

		log.Info("example requests complete — shutting down",
			tlog.String("text_log", textPath),
			tlog.String("json_log", jsonPath),
		)
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", tlog.Err(err))
	}
}
