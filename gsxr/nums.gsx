package gsxr

import "github.com/gsxhq/gsx-bench/data"

// Stats is the numeric-interpolation scenario: a table whose cells interpolate
// integers (a large computed value, a length, an id). Exercises the per-render
// numeric scratch buffer (gw.IntInto) vs templ's strconv.Itoa.
component Stats(rows []data.Row) {
	<table>{ for _, r := range rows {
		<tr><td>{ r.ID * 1009 }</td><td>{ len(r.Email) }</td><td>{ r.ID }</td></tr>
	} }</table>
}
