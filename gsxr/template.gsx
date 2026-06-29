package gsxr

import "github.com/gsxhq/gsx-bench/data"

// Render mirrors a-h/templ's benchmark template so gsx and templ render the
// same HTML for an apples-to-apples comparison. Boolean attributes use gsx's
// conditional-attribute form, the analogue of templ's `attr?={ cond }`.
component Render(p data.Person) {
	<div>
		<h1>{ p.Name }</h1>
		<div style="font-family: 'sans-serif'" id="test" data-contents={ `something with "quotes" and a <tag>` }>
			<div>email:<a href={"mailto: " + p.Email}>{ p.Email }</a></div>
		</div>
	</div>
	<hr { if true { noshade } }/>
	<hr optionA { if true { optionB } } optionC="other" { if false { optionD } }/>
	<hr noshade/>
}
