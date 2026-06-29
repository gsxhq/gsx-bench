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
| `Page` | a realistic whole page: full document, nested components, loop, conditional, text, static + dynamic (URL) + boolean attrs, and multi-token utility classes on every component root | yes |
| `Comments` | escaping-heavy: bodies dense with `< > & " '` stress the HTML text escaper | yes |
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
| Document | **266 ns · 56 B · 2 allocs** | 390 ns · 361 B · 10 | 1412 ns · 642 B · 24 |
| List (20 rows) | **1489 ns · 80 B · 2 allocs** | 3608 ns · 1912 B · 123 | — |
| Table (20 children) | **2255 ns · 1634 B · 21 allocs** | 4945 ns · 4806 B · 183 | — |
| Page (realistic, class-heavy) | **4699 ns · 2207 B · 61 allocs** | 6701 ns · 4967 B · 204 allocs | — |
| Comments (escaping-heavy) | **3690 ns · 32 B · 1 alloc** | 6508 ns · 9074 B · 143 allocs | — |
| Piped (40 filters) | 1870 ns · 400 B · 42 allocs | — | — |

gsx beats templ on every shared scenario — most dramatically on lists, where its
inline loop is **allocation-flat (2 allocs total for 20 rows)** while templ does
123, and on the realistic `Page` where gsx is faster with ~half the memory and
under a third of the allocations.

**`Page` is the headline:** a realistic, class-heavy page renders ~1.4× faster
than templ at 2207 B / 61 allocs vs 4967 B / 204. Earlier this scenario was the
one place gsx *lost* (≈10 µs / 122 allocs), because the per-component-root
attribute machinery was wasteful on every render. Two landed fixes closed it:

- **Class merge** moved into the configurable `ClassMerger` — a single class
  source is returned verbatim (no tokenize/dedup/join), a lone class token skips
  the merger entirely, and the default merge dedupes in place without a map.
- **Empty root-attr fast path** — `StyleMerged("","")` and `Attrs.Without` on an
  empty bag (the common no-fallthrough case) now return immediately instead of
  building throwaway dedup maps (12× faster on that path, 0 allocs).

Together the realistic page went from ~10 µs / 122 allocs → **4.7 µs / 61**.
`Page` was the gate for that work; it now guards against regressing it.

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

## What the profile says

Escape analysis and allocation profiling (`go test -bench … -memprofile`) on the
Pooled runs show:

- **`gsx.Writer` does not escape** — it's stack-allocated; the writer wrapper is
  free.
- **Inline loops are allocation-flat** — `List` is 2 allocs whether it renders 1
  row or 20.
- **Text escaping is allocation-free** — `Comments` (entity-dense bodies) is 1
  alloc / 32 B for gsx vs 143 / 9 KB for templ. gsx's escaper (a `strings.Replacer`
  port of `html/template`) writes safe runs straight to the output and only
  diverts for the rare entity, so escaping never allocates. A strength, not a
  target.
- **Class merging — fixed.** Multi-token utility classes on component roots used
  to dominate `Page` (`strings.Fields` + a map-based dedup + `strings.Join` per
  render) and made gsx lose. The merge now lives in the `ClassMerger`: single
  source returned verbatim, lone token skips the merger, default dedup is
  map-free. `Page` fell from ~10 µs / 122 allocs to ~5.6 µs / 62.
- **Composition allocates per child** — `Table`'s remaining allocs are the lazy
  node closure each `<Card/>` builds (props captured + boxed through `Node`).
  Removing it was prototyped (a generated carrier type + a generic render helper,
  0 alloc/child even cross-package) but **declined**: it needs an exported type
  per component and a full-corpus golden churn, and templ pays the identical
  per-component closure cost (its `Card(r).Render(...)` builds the same capturing
  closure — and templ does *more* per child, 183 allocs on this `Table`). Not a
  competitive gap; not worth the surface.
- **Pipeline filters allocate per call** — `Piped` is 42 allocs for 40 filter
  applications; a string transform like `upper` must allocate. Inherent.

These per-scenario, per-path numbers are the baseline any future write-path
change should be measured against.

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
