# tlog Benchmarks

Performance measurements for `tlog`. All results are **medians of 3 runs** on the same machine.

```
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics (12 logical cores)
os:  Linux / amd64
go:  1.26.3
```

Run the suite yourself:

```sh
go test -bench=. -benchmem -count=3 -run='^$' ./benchmarks/
```

---

## tlog

Output routed to `discard` (no I/O cost) unless noted. Measures formatter + dispatch overhead.

### Core formatters

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Info / no fields / text | 418 | 72 | 2 |
| Info / no fields / JSON | 321 | 112 | 2 |
| Info / 5 fields / text | 803 | 368 | 5 |
| Info / 5 fields / JSON | 779 | 419 | 4 |
| Info / 10 fields / text | 1 256 | 658 | 7 |
| Info / 10 fields / JSON | 1 285 | 712 | 6 |
| Debug filtered (min=Info) | 39 | 48 | 1 |
| WithCaller / text | 1 010 | 408 | 5 |
| WithCaller / JSON | 1 011 | 512 | 6 |

### Outputs

| Benchmark             | ns/op | B/op | allocs/op |
| --------------------- | ----: | ---: | --------: |
| Console output / text |   506 |  136 |         3 |
| Console output / JSON |   423 |  176 |         3 |
| File output / JSON    |   581 |  176 |         3 |

### Parallel writes (GOMAXPROCS=12)

| Benchmark                  | ns/op | B/op | allocs/op |
| -------------------------- | ----: | ---: | --------: |
| Parallel / text            |   175 |  136 |         3 |
| Parallel / JSON            |   178 |  176 |         3 |
| Parallel / pretty (styled) | 1 560 |  384 |        20 |

### Pretty formatter

Pretty is a human-facing console formatter. Its overhead is dominated by
lipgloss ANSI rendering and is not intended for high-throughput production use.

| Benchmark                    |  ns/op |  B/op | allocs/op |
| ---------------------------- | -----: | ----: | --------: |
| No fields / styled           |  4 391 |   208 |        11 |
| 5 fields / styled            | 20 151 | 1 088 |        54 |
| 5 fields / NoColor theme     |  8 986 |   376 |         6 |
| 5 fields / plain (TTY=false) |    814 |   368 |         5 |
| 10 fields / styled           | 37 305 | 1 953 |        96 |

The `NoColor` theme eliminates most lipgloss per-field overhead. The plain
(`TTY=false`) path is equivalent to the text formatter.

---

## Reference comparison

Comparison against popular Go loggers, all writing to a no-op output with 5
fields unless noted. Numbers are for context only — each library makes
different design trade-offs.

| Library         | Scenario          | ns/op | B/op | allocs/op |
| --------------- | ----------------- | ----: | ---: | --------: |
| **slog** (text) | No fields         |   645 |    0 |         0 |
| **slog** (text) | 5 fields          | 1 495 |  240 |         5 |
| **slog** (JSON) | 5 fields          | 1 406 |  240 |         5 |
| **slog**        | Filtered          |    46 |   48 |         1 |
| **slog**        | Parallel 5 fields |   229 |   48 |         1 |
| **zap** (JSON)  | No fields         |   365 |    0 |         0 |
| **zap** (JSON)  | 5 fields          |   872 |  320 |         1 |
| **zap**         | Filtered          |    56 |   64 |         1 |
| **zap**         | Parallel 5 fields |   135 |   64 |         1 |
| **zerolog**     | No fields         |    84 |    0 |         0 |
| **zerolog**     | 5 fields          |   273 |    0 |         0 |
| **zerolog**     | Filtered          |     6 |    0 |         0 |
| **zerolog**     | Parallel 5 fields |    17 |    0 |         0 |
| **tlog** (text) | No fields         |   418 |   72 |         2 |
| **tlog** (text) | 5 fields          |   803 |  368 |         5 |
| **tlog** (JSON) | 5 fields          |   779 |  419 |         4 |
| **tlog**        | Filtered          |    39 |   48 |         1 |
| **tlog**        | Parallel JSON     |   178 |  176 |         3 |

### Analysis

- **zerolog** achieves zero allocations via its builder API; it avoids the
  `[]Field` slice that key=value loggers (tlog, slog, zap) allocate per call.
- **tlog / text (5 fields)** is comparable to **slog / text** and faster than
  **slog / JSON** — the gap vs zap is primarily the `[]Field` allocation.
- **tlog** filtered overhead (39 ns) is on par with slog/zap (~46/56 ns).
- **tlog** parallel throughput (178 ns JSON) is between slog (229 ns) and zap
  (135 ns).

### Allocation baseline

The `Info` hot path with the text formatter allocates **2 allocs/op** at
72 B/op for the no-fields case. These come from:

1. The `[]Field` slice header (even when empty, the variadic call allocates).
2. The formatted byte slice returned by the formatter.

This is the documented acceptable baseline for tlog's design. Zero-allocation
logging would require a builder API (zerolog style), which is out of scope for
v1.
