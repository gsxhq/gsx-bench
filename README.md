# gsx-bench

Runtime render benchmarks comparing [**gsx**](https://github.com/gsxhq/gsx) against
[**a-h/templ**](https://github.com/a-h/templ), Go's `html/template`, and a raw
`io.WriteString` floor.

This lives in its own module on purpose: gsx's runtime is standard-library-only,
and we don't want templ (or the templ/gsx code generators) anywhere near gsx's
`go.mod`. Here, gsx is wired in via a `replace` directive to the local working
tree (`../gsx`), so the numbers track `main` as it evolves; templ is pulled from
its **published module** (origin main), never a local experimental checkout.

## What it renders

All engines render the same document — a faithful port of templ's own benchmark
template (`benchmarks/templ/template.templ`): nested elements, text
interpolation, a `mailto:` URL attribute, an attribute value with characters
that must be HTML-escaped, and boolean/void attributes.

`TestRenderersAgree` asserts gsx and templ produce the same HTML, modulo two
cosmetic, browser-irrelevant differences (documented in `render_test.go`):

- **Void elements** — gsx emits `<hr/>`, templ emits `<hr>` (the trailing slash
  is ignored by HTML5 parsers).
- **Attribute escaping** — inside a double-quoted value gsx escapes `'` to
  `&#39;`; templ leaves the literal. gsx's escaper is a faithful port of
  `html/template`, which escapes it the same way (see the `goTemplate` reference).

## Running

```sh
go test -bench . -benchmem -run '^$' .   # or: make bench
go test -v -run TestRenderersAgree .     # or: make test
```

## Regenerating

Generated `.go` files are committed so the benchmark runs out of the box. To
regenerate after editing a template:

```sh
make tools       # one-time: installs the standalone templ CLI
make generate    # gsx (run from ../gsx) + templ
```

## Results

Apple M3 Ultra, Go 1.26.1, gsx @ `main`, templ v0.3.1020 (`-count=6`, representative):

```
BenchmarkGSX-32              ~423 ns/op     784 B/op      8 allocs/op
BenchmarkTempl-32           ~448 ns/op     650 B/op     11 allocs/op
BenchmarkGoTemplate-32     ~1520 ns/op    1266 B/op     28 allocs/op
BenchmarkIOWriteString-32    ~48 ns/op     320 B/op      1 allocs/op
```

**Takeaway:** gsx and templ are neck-and-neck on render time — both ~3.5× faster
than `html/template`. gsx does fewer allocations (8 vs 11) at slightly more bytes
per op; templ allocates a bit less memory. The `io.WriteString` row is the
no-templating floor (a single write of a precomputed string).

Numbers are machine- and version-specific — run it yourself for figures that
mean something on your hardware.
