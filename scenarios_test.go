package bench

import (
	"testing"

	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

// Feature scenarios beyond Document, each exercising a distinct writer path.
// Benchmarked on the production-realistic Pooled destination.

// List — for-range loop + repeated text + inline if (control flow). gsx's inline
// loop is allocation-flat: 2 allocs total regardless of row count.
func BenchmarkListGSXPooled(b *testing.B) {
	pooled(b, gsxRender(gsxr.List(gsxr.ListProps{Rows: rows})))
}
func BenchmarkListTemplPooled(b *testing.B) { pooled(b, templRender(templr.List(rows))) }

// Table — component composition: a Card child rendered per row. Exercises the
// child-Node + props path, which allocates per child (the hot spot to watch).
func BenchmarkTableGSXPooled(b *testing.B) {
	pooled(b, gsxRender(gsxr.Table(gsxr.TableProps{Rows: rows})))
}
func BenchmarkTableTemplPooled(b *testing.B) { pooled(b, templRender(templr.Table(rows))) }

// Piped — the pipeline (|>) filter call path. gsx-only (templ has no |>).
func BenchmarkPipedGSXPooled(b *testing.B) {
	pooled(b, gsxRender(gsxr.Piped(gsxr.PipedProps{Rows: rows})))
}

// Page — the realistic "whole page": nested components, loop, conditional, text,
// static+dynamic+boolean attrs, and heavy multi-token utility classes through
// the class-merge path. The closest scenario to a real server render.
func BenchmarkPageGSXPooled(b *testing.B) {
	pooled(b, gsxRender(gsxr.Page(gsxr.PageProps{Rows: rows})))
}
func BenchmarkPageTemplPooled(b *testing.B) { pooled(b, templRender(templr.Page(rows))) }

// Comments — escaping-heavy: bodies full of <, >, &, ", ' stress the HTML text
// escaper (gsx's strings.Replacer port of html/template) vs templ's escaper.
func BenchmarkCommentsGSXPooled(b *testing.B) {
	pooled(b, gsxRender(gsxr.Comments(gsxr.CommentsProps{Items: comments})))
}
func BenchmarkCommentsTemplPooled(b *testing.B) { pooled(b, templRender(templr.Comments(comments))) }
