# tlog examples

Runnable programs demonstrating tlog features from basic usage to production patterns.

## Interactive menu

```bash
go run ./examples/
```

An interactive menu (built with [lipgloss](https://github.com/charmbracelet/lipgloss)) lets you pick and launch any example. Press `q` or `Ctrl+C` to quit.

Non-interactive use (e.g. CI):

```bash
echo q | go run ./examples/
```

## Examples

| #   | Directory                | Description                                                      | Run command                       |
| --- | ------------------------ | ---------------------------------------------------------------- | --------------------------------- |
| 1   | `examples/basic/`        | Console output, all log levels, basic fields                     | `go run ./examples/basic/`        |
| 2   | `examples/structured/`   | All nine field constructors, nil error, child logger             | `go run ./examples/structured/`   |
| 3   | `examples/multi-output/` | Console + file fan-out, append across sessions                   | `go run ./examples/multi-output/` |
| 4   | `examples/custom-theme/` | All nine built-in themes, `WithStyles`, fully custom `Styles`    | `go run ./examples/custom-theme/` |
| 5   | `examples/json-mode/`    | Production JSON logger, RFC 3339 Nano, per-request child loggers | `go run ./examples/json-mode/`    |
| 6   | `examples/context/`      | `NewContext` / `FromContext` / `WithContext`, middleware pattern | `go run ./examples/context/`      |
| 7   | `examples/global/`       | `SetDefault`, package-level helpers, runtime logger swap         | `go run ./examples/global/`       |
| 8   | `examples/webserver/`    | HTTP access log, panic recovery, dual-formatter multi-output     | `go run ./examples/webserver/`    |

## webserver outputs

The webserver example fans log entries to **four simultaneous destinations** via two loggers:

```
┌─────────────────────────────────┬────────────────────────────────────────────┐
│ Logger                          │ Outputs                                    │
├─────────────────────────────────┼────────────────────────────────────────────┤
│ humanLog  (text/pretty format)  │ stderr — pretty coloured console           │
│                                 │ $TMPDIR/tlog-webserver.log — plain text    │
├─────────────────────────────────┼────────────────────────────────────────────┤
│ machLog   (JSON format)         │ stdout — JSON stream                       │
│                                 │ $TMPDIR/tlog-webserver.json — JSON file    │
└─────────────────────────────────┴────────────────────────────────────────────┘
```

Because tlog uses **one formatter per logger**, human-readable and machine-readable
outputs require two separate loggers. The `dualLog` wrapper in the example delegates
every log call to both, keeping handler code format-agnostic.

File paths are printed on startup. Inspect them after the run:

```bash
cat /tmp/tlog-webserver.log   # plain text
cat /tmp/tlog-webserver.json  # newline-delimited JSON
```
