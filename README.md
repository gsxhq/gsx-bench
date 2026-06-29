# gsx-bench

Runtime render benchmarks for [**gsx**](https://github.com/gsxhq/gsx), compared against [**a-h/templ**](https://github.com/a-h/templ), Go's `html/template`, and a raw `io.WriteString` floor.

## Two axes

The suite varies two things independently:

**Scenario** — which template feature is exercised. Different features hit
different writer paths, so per-scenario numbers are what tell you an optimisation
helped the path it targeted without regressing the others.

| scenario | exercises | templ counterpart |
| --- | --- | --- |
| `Document` | nested elements, text, a `mailto:` URL attr, an escape-requiring attr value, void/boolean attrs (a faithful port of templ's own benchmark template) | yes |
| `List` | `for`-range loop + repeated text + inline `if`, over 20 rows | yes |
| `Table` | component composition: a `Card` child rendered once per row | yes |
| `Piped` | the pipeline (`|>`) filter call path | gsx-only (templ has no `|>`) |

**Destination** — where the bytes go. This matters more than it looks:

| destination | what it models |
| --- | --- |
| `Pooled` | **production-realistic.** A warm `*bytes.Buffer` from a `sync.Pool`, one per request — mirrors structpages' buffered http middleware. `bytes.Buffer.Reset()` keeps its backing array, so the buffer never reallocates across requests. |
| `Discard` | pure engine overhead — `io.Discard` removes all destination cost. |
| `Builder` | the cold destination templ's *own* benchmark uses: a `strings.Builder` whose `Reset()` nils the backing array, forcing it to regrow from scratch every iteration. Useful for cross-referencing published templ numbers; **not** representative of real use. |

`TestScenariosAgree` pins gsx and templ to byte-identical output for every shared
scenario (modulo two documented cosmetic deltas — see `canonical` in
`bench_test.go`), so the comparison is honest.

## Results

Apple M3 Ultra, Go 1.26.1, gsx @ `main`, templ v0.3.1020.

### Production-realistic (Pooled buffer)

| scenario | gsx | templ | html/template |
| --- | --- | --- | --- |
| Document | **272 ns · 56 B · 2 allocs** | 379 ns · 361 B · 10 | 1418 ns · 642 B · 24 |
| List (20 rows) | **1505 ns · 80 B · 2 allocs** | 3565 ns · 1912 B · 123 | — |
| Table (20 children) | **4475 ns · 2323 B · 62 allocs** | 4816 ns · 4806 B · 183 | — |
| Piped (40 filters) | 1920 ns · 400 B · 42 allocs | — | — |

With the buffer a real consumer uses, gsx beats templ on every shared scenario —
most dramatically on lists, where gsx's inline loop is **allocation-flat (2
allocs total for 20 rows)** while templ does 123.

### Document, across destinations

This is the same template rendered three ways — it shows how much of a "render
benchmark" is actually the *destination*, not the engine:

| destination | gsx | templ | html/template | raw write |
| --- | --- | --- | --- | --- |
| Pooled (warm buffer) | **272 ns · 2 allocs** | 379 ns · 10 | 1418 ns · 24 | — |
| Discard (engine only) | **186 ns · 2 allocs** | 372 ns · 10 | — | — |
| Builder (cold, templ-style) | 422 ns · 8 allocs | 417 ns · 11 | 1522 ns · 28 | 53 ns · 1 |

Note the **Builder row is a near-tie** (422 vs 417 ns) — and it's the one most
benchmarks publish. That tie is an artifact: gsx streams many small writes
straight to the destination, so against a `strings.Builder` that nils itself
every iteration it pays for ~6 buffer regrowths. Move to a warm buffer (Pooled)
or remove the destination (Discard) and gsx's real advantage appears: **2× faster
than templ on pure engine overhead, with 5× fewer allocations.**

The architectural difference behind it: gsx renders by streaming fragments
directly to your `io.Writer` (intrinsic cost: a stack-allocated writer + the lazy
node closure); templ renders into its own pooled buffer and flushes once (more
intrinsic allocations, fewer destination writes).

## What the profile says (optimisation targets)

Escape analysis and allocation profiling (`go test -bench … -memprofile`) on the
Pooled runs show:

- **`gsx.Writer` does not escape** — it's stack-allocated; the writer wrapper is
  already free.
- **Inline loops are allocation-flat** — `List` is 2 allocs whether it renders 1
  row or 20. This path is already optimal.
- **Composition allocates per child** — `Table` is 62 allocs for 20 cards (~3
  each): each `<Card r={r}/>` builds a props value and a node closure that
  escapes through the `Writer.Node` path. This is the main lever.
- **Pipeline filters allocate per call** — `Piped` is 42 allocs for 40 filter
  applications; string-transform filters (`upper`) allocate a new string each
  call. Partly inherent, partly improvable.

These per-scenario, per-path numbers are the baseline any write-path optimisation
should be measured against.

## Running

```sh
make bench                          # full sweep, -benchmem
make test                           # output-equivalence (gsx == templ)
go test -bench Pooled -benchmem -run '^$' .   # just the realistic numbers
go test -bench Table  -benchmem -run '^$' .   # one scenario, all engines
```

Numbers are machine- and version-specific — run it yourself for figures that mean
something on your hardware.

## Regenerating

Generated `.go` files are committed so the benchmark runs out of the box:

```sh
make tools       # one-time: installs the standalone templ CLI
make generate    # gsx (run from ../gsx) + templ
```
