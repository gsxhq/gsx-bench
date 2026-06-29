package bench

import (
	"context"
	"html/template"
	"io"
	"strings"
	"testing"

	"github.com/gsxhq/gsx-bench/data"
	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

var person = data.Person{
	Name:  "Luiz Bonfa",
	Email: "luiz@example.com",
}

// goTemplate is the equivalent html/template, copied from a-h/templ's own
// benchmark so all three template engines render the same document.
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

func gsxOut() string {
	var b strings.Builder
	if err := gsxr.Render(gsxr.RenderProps{P: person}).Render(context.Background(), &b); err != nil {
		panic(err)
	}
	return b.String()
}

func templOut() string {
	var b strings.Builder
	if err := templr.Render(person).Render(context.Background(), &b); err != nil {
		panic(err)
	}
	return b.String()
}

// canonical removes two cosmetic, browser-irrelevant differences between the
// engines so the comparison checks that they render the *same document*:
//
//   - Void elements: gsx emits `<hr/>` (a trailing slash HTML5 parsers ignore),
//     templ emits `<hr>`.
//   - Attribute escaping: inside a double-quoted value gsx escapes `'` to
//     `&#39;` (its escaper is a faithful port of html/template, whose own output
//     — see goTemplate below — escapes it too); templ leaves the literal quote.
//
// Any *other* divergence still fails the test, keeping the benchmark honest.
func canonical(s string) string {
	s = strings.ReplaceAll(s, "/>", ">")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}

// TestRenderersAgree confirms gsx and templ render the same document, so the
// benchmark below is an apples-to-apples comparison rather than two engines
// emitting structurally different HTML.
func TestRenderersAgree(t *testing.T) {
	g, tp := gsxOut(), templOut()
	if canonical(g) != canonical(tp) {
		t.Fatalf("gsx and templ output differ beyond known cosmetic deltas:\n gsx:   %q\n templ: %q", g, tp)
	}
	t.Logf("gsx HTML   (%d bytes): %s", len(g), g)
	t.Logf("templ HTML (%d bytes): %s", len(tp), tp)
}

func BenchmarkGSX(b *testing.B) {
	b.ReportAllocs()
	n := gsxr.Render(gsxr.RenderProps{P: person})
	w := new(strings.Builder)
	ctx := context.Background()
	for range b.N {
		if err := n.Render(ctx, w); err != nil {
			b.Fatal(err)
		}
		w.Reset()
	}
}

func BenchmarkTempl(b *testing.B) {
	b.ReportAllocs()
	c := templr.Render(person)
	w := new(strings.Builder)
	ctx := context.Background()
	for range b.N {
		if err := c.Render(ctx, w); err != nil {
			b.Fatal(err)
		}
		w.Reset()
	}
}

func BenchmarkGoTemplate(b *testing.B) {
	b.ReportAllocs()
	w := new(strings.Builder)
	for range b.N {
		if err := goTemplate.Execute(w, person); err != nil {
			b.Fatal(err)
		}
		w.Reset()
	}
}

const rawHTML = `<div><h1>Luiz Bonfa</h1><div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;"><div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div></div></div><hr noshade><hr optionA optionB optionC="other"><hr noshade>`

func BenchmarkIOWriteString(b *testing.B) {
	b.ReportAllocs()
	w := new(strings.Builder)
	for range b.N {
		if _, err := io.WriteString(w, rawHTML); err != nil {
			b.Fatal(err)
		}
		w.Reset()
	}
}
