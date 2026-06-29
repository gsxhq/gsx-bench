package gsxr

import "github.com/gsxhq/gsx-bench/data"

// Comments is the escaping-heavy scenario: each comment body carries many
// characters that must be HTML-escaped, stressing the text escaper.
component Comments(items []data.Comment) {
	<section>{ for _, c := range items {
		<article><b>{ c.Author }</b>: { c.Body }</article>
	} }</section>
}
