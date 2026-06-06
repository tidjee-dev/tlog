# `tlog`

A lightweight, extensible, production-oriented logger for Go.

> Minimal API. Beautiful output. Zero magic.

## Features

- Minimal, expressive public API
- Structured logging with typed fields
- Human-readable console output with [lipgloss](https://github.com/charmbracelet/lipgloss) styling
- JSON output compatible with ELK / Loki
- Multiple simultaneous outputs (console, file, discard)
- Functional options pattern for clean configuration
- Fully customizable themes and styles
- Thread-safe by design
- Immutable child loggers — `With` and `WithLevel` for scoped logging
- Context-aware API — store and retrieve loggers from `context.Context`
- Optional global logger — safe package-level helpers, opt-in only
- No global mutable state by default

## Installation

```sh
go get github.com/tidjee-dev/tlog
```

## Quick Start

```go
package main

import (
    "errors"
    "time"

    "github.com/tidjee-dev/tlog"
    "github.com/tidjee-dev/tlog/level"
)

func main() {
    log := tlog.New(
        tlog.WithLevel(level.DEBUG),
        tlog.WithConsole(), // pretty on TTY, plain text otherwise
    )

    log.Info("server started",
        tlog.String("host", "localhost"),
        tlog.Int("port", 8080),
    )

    log.Warn("slow query",
        tlog.Duration("latency", 322*time.Millisecond),
    )

    log.Error("db connection failed",
        tlog.Err(errors.New("dial tcp: connection refused")),
        tlog.Int("retry", 3),
    )
}
```

### Console output (text formatter)

```
2026-05-31T20:14:33.000 INFO  server started  host=localhost  port=8080
2026-05-31T20:14:34.000 WARN  slow query  latency=322ms
2026-05-31T20:14:35.000 ERROR db connection failed  error="dial tcp: connection refused"  retry=3
```

### JSON output

Swap to the JSON formatter with a single option:

```go
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithJSON(),
)
```

```json
{"time":"2026-05-31T20:14:33+02:00","level":"INFO","msg":"server started","host":"localhost","port":8080}
{"time":"2026-05-31T20:14:34+02:00","level":"WARN","msg":"slow query","latency":"322ms"}
{"time":"2026-05-31T20:14:35+02:00","level":"ERROR","msg":"db connection failed","error":"dial tcp: connection refused","retry":3}
```

## Log Levels

| Level   | Description                         |
| ------- | ----------------------------------- |
| `TRACE` | Extremely verbose diagnostic info   |
| `DEBUG` | Development-time diagnostics        |
| `INFO`  | Normal operational messages         |
| `WARN`  | Non-critical issues worth noting    |
| `ERROR` | Errors that need attention          |
| `FATAL` | Critical failure, exits the program |
| `PANIC` | Critical failure, panics            |

## Structured Fields

```go
tlog.String("key", "value")
tlog.Int("port", 8080)
tlog.Float64("ratio", 0.95)
tlog.Bool("enabled", true)
tlog.Duration("latency", d)
tlog.Time("started_at", t)
tlog.Err(err)
tlog.Any("meta", someStruct)
```

## Configuration

```go
log := tlog.New(
    tlog.WithLevel(level.INFO),           // minimum log level
    tlog.WithConsole(),                   // enable console output
    tlog.WithFile("app.log"),             // enable file output
    tlog.WithJSON(),                      // use JSON formatter (default: text)
    tlog.WithCaller(),                    // include caller info
    tlog.WithTimestampFormat(time.RFC3339),
)
defer log.Close() // flush and release file handles
```

## Child Loggers

Create scoped, immutable child loggers that inherit the parent's configuration.
The parent is never mutated.

```go
// With — prepend fields to every entry from this logger and its children.
base := tlog.New(tlog.WithConsole())

svc := base.With(
    tlog.String("svc", "orders"),
    tlog.String("region", "eu-west-1"),
)
svc.Info("service ready")

// Each request gets its own child with a request ID.
req := svc.With(tlog.String("request_id", "req-42"))
req.Info("checkout started", tlog.String("method", "POST"))
// → svc=orders  region=eu-west-1  request_id=req-42  method=POST

// WithLevel — override the minimum level for a child; parent is unaffected.
prod := base.
    With(tlog.String("env", "production")).
    WithLevel(level.Error)   // silence everything below Error

prod.Debug("filtered — not emitted")
prod.Error("disk full", tlog.Err(err))

base.Debug("base still logs at its own level")
```

## Context Support

Store a logger in a `context.Context` and retrieve it deep in the call stack
without threading it through every function signature.

```go
import "context"

// Attach a logger to a request context.
log := tlog.New(tlog.WithConsole()).With(tlog.String("request_id", "req-abc"))
ctx := tlog.NewContext(r.Context(), log)

// Deep in the call stack — no logger argument needed.
func handleOrder(ctx context.Context, orderID string) {
    log := tlog.FromContext(ctx)   // never nil — returns a no-op if absent
    log.Info("processing order", tlog.String("order", orderID))
}

// Add fields to the logger inside the context without replacing it.
ctx = tlog.WithContext(ctx, tlog.String("trace_id", "tr-xyz"))
tlog.FromContext(ctx).Info("db query complete", tlog.Duration("took", 3*time.Millisecond))
// → request_id=req-abc  trace_id=tr-xyz  took=3ms
```

## Global Logger

An optional, opt-in package-level logger. Safe for applications; not recommended
for libraries.

```go
// Default is a no-op — nothing prints until you call SetDefault.
tlog.Info("silently dropped")

// Replace the default once at startup.
tlog.SetDefault(tlog.New(
    tlog.WithConsole(),
    tlog.WithLevel(level.Info),
))

// Use from anywhere without passing a logger around.
tlog.Info("server started", tlog.String("host", "0.0.0.0"), tlog.Int("port", 8080))
tlog.Warn("high memory", tlog.Float64("pct", 91.2))
tlog.Error("upstream timeout", tlog.Err(err))

// SetDefault(nil) panics — prevents silent nil-logger bugs.
```

## Styling

`tlog` auto-selects the pretty formatter when the console is a TTY and the
plain-text formatter when it is not (e.g. CI, file redirect). Override with
`WithTheme`, `WithStyles`, or `WithPretty`:

```go
import (
    "github.com/tidjee-dev/tlog"
    "github.com/tidjee-dev/tlog/styles"
)

// Pick a built-in theme.
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithTheme(styles.Dev()),
)

// Mutate individual fields of the active theme.
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithTheme(styles.Minimal()),
    tlog.WithStyles(func(s *tlog.Styles) {
        s.Message = s.Message.Bold(true)
    }),
)

// Force pretty output even when stdout is not a TTY.
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithPretty(),
)

// Web-server access logging with badge level pills.
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithTheme(styles.BadgyHTTP()),
)
log.Info("request",
    tlog.String("method", "GET"),
    tlog.String("path", "/api/v1/users"),
    tlog.Int("status", 200),
    tlog.Duration("latency", 4*time.Millisecond),
)
```

You can also reach deeper via the `console` package directly:

```go
import "github.com/tidjee-dev/tlog/outputs/console"

// Force plain-text (no ANSI) regardless of terminal.
log := tlog.New(
    tlog.WithConsole(console.WithTTY(false)),
)
```

### Built-in Themes

| Theme        | Description                                         |
| ------------ | --------------------------------------------------- |
| `Dev`        | High contrast, verbose dev mode (default)           |
| `Minimal`    | Clean, low visual noise                             |
| `Monochrome` | Bold/italic only, no color                          |
| `NoColor`    | No ANSI codes — CI / file output                    |
| `Production` | Subdued, ops-oriented                               |
| `Badgy`      | Badge-style level pills with background color       |
| `HTTP`       | Web-server palette: cyan keys, warm-yellow values   |
| `BadgyHTTP`  | Badge levels + HTTP palette — ideal for access logs |

## Outputs

### Console

```go
tlog.WithConsole()
```

Auto-detects TTY. Uses the `pretty` formatter (lipgloss-styled) when running in
a terminal, and the plain-text formatter otherwise. Pass `console.WithStyles`
to choose a theme, or `console.WithTTY(false)` to force plain text.

### File

```go
log := tlog.New(
    tlog.WithConsole(),
    tlog.WithFile("app.log"),
)
defer log.Close() // flush buffer and close the file
```

Opens (or creates) the file for appending. Writes are buffered; call `Close()` to
flush and release the file handle.

### Discard

```go
tlog.WithDiscard()
```

Silently drops all log entries. Useful in tests and benchmarks where output is not needed.

## Performance

Benchmarks run on AMD Ryzen 5 PRO 4650G · Go 1.26.3 · `go test -bench=. -benchmem ./benchmarks/`

| Scenario                           |  ns/op |  B/op | allocs/op |
| ---------------------------------- | -----: | ----: | --------: |
| `Info` no fields / text            |    528 |    88 |         3 |
| `Info` no fields / JSON            |    322 |   112 |         2 |
| `Info` 5 fields / text             |  1 470 |   388 |         7 |
| `Info` 5 fields / JSON             |    910 |   419 |         4 |
| `Info` 10 fields / text            |  2 022 |   684 |        11 |
| `Info` 10 fields / JSON            |  1 328 |   712 |         6 |
| `Debug` filtered (level=Info)      |     42 |    48 |         1 |
| `Info` + caller / JSON             |  1 015 |   512 |         6 |
| Parallel `Info` / JSON             |    170 |   176 |         3 |
| `Info` 5 fields / pretty (styled)  | 27 132 | 2 498 |       116 |
| `Info` 5 fields / pretty (NoColor) |  9 460 |   392 |         8 |

See [`benchmarks/`](benchmarks/) for the full suite including reference comparisons
against `slog`, `zap`, and `zerolog`.

## Architecture

```
Logger
  ↓
Entry  (immutable log event)
  ↓
Formatter / Encoder  (text | json | pretty)
  ↓
Output(s)  (console | file | discard)
  ↓
Destination
```

### Package layout

```
tlog/
├── core/          Logger, Config, Context
├── level/         Level type and parser
├── interfaces/    Entry, Field, Formatter, Output, Encoder, Style
├── formatter/     text/, json/, pretty/
├── outputs/       console/, file/, discard/
├── styles/        Default, Minimal, Monochrome, NoColor, Dev, Production,
│                  Badgy, HTTP, BadgyHTTP themes
├── internal/      buffer, clock, syncpool, writer (unexported)
├── pkg/ansi/      ANSI helpers (exported utility)
├── examples/
├── benchmarks/
└── tests/
```

## Error Handling

`tlog` never panics internally (except on `Fatal` and `Panic` log calls).

Output errors are handled silently by default or via a custom handler:

```go
log := tlog.New(
    tlog.WithErrorHandler(func(err error) {
        fmt.Fprintln(os.Stderr, "tlog internal error:", err)
    }),
)
```

## Design Principles

- Small, composable interfaces
- Immutable log entries and immutable child loggers
- Typed fields — no `map[string]interface{}`
- Explicit dependencies, no hidden goroutines
- Context-aware API — logger stored/retrieved from `context.Context`
- Optional global logger — opt-in, no forced package-level state
- Safe concurrency — `sync.Pool` for buffer reuse, `sync/atomic` for default logger
- Filtered entries skip formatting entirely — near-zero overhead

## Inspiration

| Project    | Influence             |
| ---------- | --------------------- |
| `zap`      | Performance approach  |
| `zerolog`  | Minimalism            |
| `slog`     | Structured API design |
| `lipgloss` | Terminal UX           |

## License

MIT
