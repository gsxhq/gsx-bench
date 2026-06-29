package bench

import (
	"io"
	"testing"

	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

// Document — a faithful port of templ's own benchmark template (nested elements,
// text interpolation, a mailto: URL attribute, an escape-requiring attribute
// value, void/boolean attributes). Benchmarked across all three destinations to
// show how much of the headline number is the destination vs the engine.

func docGSX() render   { return gsxRender(gsxr.Render(gsxr.RenderProps{P: person})) }
func docTempl() render { return templRender(templr.Render(person)) }
func docGoTmpl() render {
	return func(w io.Writer) error { return goTemplate.Execute(w, person) }
}

func BenchmarkDocumentGSXPooled(b *testing.B)        { pooled(b, docGSX()) }
func BenchmarkDocumentTemplPooled(b *testing.B)      { pooled(b, docTempl()) }
func BenchmarkDocumentGoTemplatePooled(b *testing.B) { pooled(b, docGoTmpl()) }

func BenchmarkDocumentGSXDiscard(b *testing.B)   { discard(b, docGSX()) }
func BenchmarkDocumentTemplDiscard(b *testing.B) { discard(b, docTempl()) }

func BenchmarkDocumentGSXBuilder(b *testing.B)        { builder(b, docGSX()) }
func BenchmarkDocumentTemplBuilder(b *testing.B)      { builder(b, docTempl()) }
func BenchmarkDocumentGoTemplateBuilder(b *testing.B) { builder(b, docGoTmpl()) }

// rawHTML is the precomputed Document output — the no-templating floor.
const rawHTML = `<div><h1>Luiz Bonfa</h1><div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;"><div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div></div></div><hr noshade><hr optionA optionB optionC="other"><hr noshade>`

func BenchmarkDocumentRawBuilder(b *testing.B) {
	builder(b, func(w io.Writer) error { _, err := io.WriteString(w, rawHTML); return err })
}
