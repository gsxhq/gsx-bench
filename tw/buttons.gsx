package tw

// Button carries base utility classes; the caller's fallthrough class can
// conflict (px-8 vs px-4, bg-red vs bg-blue) and the configured merger (the
// mock Tailwind-style merger) resolves it.
component Button(label string) {
	<button class="px-4 py-2 bg-blue-500 text-white rounded font-medium" { attrs... }>
		{ label }
	</button>
}

component Buttons(labels []string, override string) {
	<div { attrs... }>
		{ for _, l := range labels {
			<Button label={l} class={ override }/>
		} }
	</div>
}
