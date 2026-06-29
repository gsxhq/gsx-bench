package bench

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/gsxhq/gsx"
	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

// Concurrency: render the realistic Page from many goroutines through the shared
// pooled buffer — a real server's load. Confirms gsx scales (no pool/alloc
// bottleneck) and keeps its single-threaded lead under contention.

func parallelGSX(b *testing.B, n gsx.Node) {
	b.ReportAllocs()
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bufPool.Get().(*bytes.Buffer)
			if err := n.Render(ctx, buf); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
			bufPool.Put(buf)
		}
	})
}

func parallelTempl(b *testing.B, c templ.Component) {
	b.ReportAllocs()
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bufPool.Get().(*bytes.Buffer)
			if err := c.Render(ctx, buf); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
			bufPool.Put(buf)
		}
	})
}

func BenchmarkPageGSXParallel(b *testing.B)   { parallelGSX(b, gsxr.Page(gsxr.PageProps{Rows: rows})) }
func BenchmarkPageTemplParallel(b *testing.B) { parallelTempl(b, templr.Page(rows)) }
