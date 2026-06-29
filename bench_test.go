// Package bench compares gsx's render performance against a-h/templ,
// html/template, and a raw io.WriteString floor.
//
// Two axes:
//
//   - Scenario — which template feature is exercised (Document, List, Table,
//     Piped). Different features hit different writer paths, so the per-scenario
//     numbers are what give confidence that an optimisation helped the path it
//     targets without regressing others.
//   - Destination — where bytes go. Pooled (a warm *bytes.Buffer from a
//     sync.Pool) is the production-realistic case, mirroring structpages'
//     buffered http middleware. Discard isolates pure engine overhead. Builder
//     (a strings.Builder that Resets to nil each iteration) is the cold
//     destination templ's own benchmark uses — useful for cross-referencing but
//     not representative.
//
// Run: go test -bench . -benchmem -run '^$'
package bench

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/templ"
	"github.com/gsxhq/gsx"
	"github.com/gsxhq/gsx-bench/data"
	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

var (
	person = data.Person{Name: "Luiz Bonfa", Email: "luiz@example.com"}
	rows   = data.Rows(20)
)

// goTemplate is the html/template equivalent of the Document scenario, copied
// from a-h/templ's own benchmark so all three engines render the same document.
var goTemplate = template.Must(template.New("example").Parse(`<div>
	<h1>{{.Name}}</h1>
	<div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;">
		<div>
			email:<a href="mailto: {{.Email}}">{{.Email}}</a></div>
		</div>
	</div>
	<hr noshade>
	<hr optionA optionB optionC="other">
	<hr noshade>
`))

// render is a destination-agnostic render thunk; one is created per benchmark
// (outside the loop) so the per-iteration cost is purely the render.
type render func(w io.Writer) error

func gsxRender(n gsx.Node) render {
	return func(w io.Writer) error { return n.Render(context.Background(), w) }
}
func templRender(c templ.Component) render {
	return func(w io.Writer) error { return c.Render(context.Background(), w) }
}

// --- destinations ---------------------------------------------------------

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// pooled mirrors a real request: draw a warm buffer from the pool, render, (a
// real handler flushes here), Reset, return. bytes.Buffer.Reset keeps its
// backing array, so the buffer never reallocates across requests.
func pooled(b *testing.B, fn render) {
	b.ReportAllocs()
	for range b.N {
		buf := bufPool.Get().(*bytes.Buffer)
		if err := fn(buf); err != nil {
			b.Fatal(err)
		}
		buf.Reset()
		bufPool.Put(buf)
	}
}

// discard removes all destination cost, leaving pure engine overhead.
func discard(b *testing.B, fn render) {
	b.ReportAllocs()
	for range b.N {
		if err := fn(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// builder is the cold destination templ's own benchmark uses: a strings.Builder
// whose Reset nils the backing array, so it regrows from scratch every iteration.
func builder(b *testing.B, fn render) {
	b.ReportAllocs()
	w := new(strings.Builder)
	for range b.N {
		if err := fn(w); err != nil {
			b.Fatal(err)
		}
		w.Reset()
	}
}

// --- output equivalence ---------------------------------------------------

// canonical removes two cosmetic, browser-irrelevant differences so the
// comparison checks the engines render the *same document*:
//
//   - Void elements: gsx emits `<hr/>`, templ emits `<hr>` (the trailing slash
//     is ignored by HTML5 parsers).
//   - Attribute escaping: inside a double-quoted value gsx escapes `'` to
//     `&#39;` (its escaper is a faithful port of html/template, whose own output
//     escapes it too); templ leaves the literal quote.
//
// Any *other* divergence still fails TestScenariosAgree.
func canonical(s string) string {
	s = strings.ReplaceAll(s, "/>", ">")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}

func renderString(fn render) string {
	var b strings.Builder
	if err := fn(&b); err != nil {
		panic(err)
	}
	return b.String()
}

// scenarios pairs each shared feature scenario's gsx and templ renderer.
var scenarios = []struct {
	name       string
	gsx, templ render
}{
	{"Document", gsxRender(gsxr.Render(gsxr.RenderProps{P: person})), templRender(templr.Render(person))},
	{"List", gsxRender(gsxr.List(gsxr.ListProps{Rows: rows})), templRender(templr.List(rows))},
	{"Table", gsxRender(gsxr.Table(gsxr.TableProps{Rows: rows})), templRender(templr.Table(rows))},
}

func TestScenariosAgree(t *testing.T) {
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			g, tp := renderString(s.gsx), renderString(s.templ)
			if canonical(g) != canonical(tp) {
				t.Fatalf("%s: gsx and templ differ beyond known cosmetic deltas:\n gsx:   %q\n templ: %q", s.name, g, tp)
			}
		})
	}
}
