package gsxr

import "github.com/gsxhq/gsx-bench/data"

// List exercises the for-range control-flow path, repeated text writes, and an
// inline if (conditional control flow).
component List(rows []data.Row) {
	<table>{ for _, r := range rows {
		<tr><td>{ r.Name }</td><td>{ r.Email }</td><td>{ r.Role }{ if !r.Active {
			<em>(inactive)</em>
		} }</td></tr>
	} }</table>
}

// Card is a leaf component; Table composes it once per row, exercising the
// component-composition path (child Node + props struct).
component Card(r data.Row) {
	<div class="card"><h3>{ r.Name }</h3><p>{ r.Email } — { r.Role }</p></div>
}

component Table(rows []data.Row) {
	<section>{ for _, r := range rows {
		<Card r={r}/>
	} }</section>
}

// Piped exercises the pipeline (|>) filter call path. gsx-only: templ has no
// pipeline operator (the equivalent is a plain function call).
component Piped(rows []data.Row) {
	<ul>{ for _, r := range rows {
		<li>{ r.Name |> upper } · { r.Role |> upper }</li>
	} }</ul>
}
