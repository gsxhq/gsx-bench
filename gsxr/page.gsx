package gsxr

import "github.com/gsxhq/gsx-bench/data"

// Page is a realistic "whole page" that exercises most features at once: a full
// document, nested component composition (Page -> UserCard), a loop, a
// conditional, text interpolation, static + dynamic (URL) + boolean attributes,
// and — on every component root — multi-token utility classes that run through
// the class-merge path. It's the closest scenario to a real server render.
component Page(rows []data.Row) {
	<html lang="en"><head><meta charset="utf-8"/><title>Users</title></head><body class="bg-gray-50 text-gray-900">
		<header class="border-b bg-white px-6 py-4"><h1 class="text-xl font-semibold tracking-tight">Users</h1><nav class="mt-2 flex gap-4"><a class="text-blue-600 hover:underline" href="/users">All</a><a class="text-blue-600 hover:underline" href="/users/active">Active</a></nav></header>
		<main class="mx-auto max-w-3xl px-6 py-4"><p class="mb-4 text-sm text-gray-500">{ len(rows) } users</p><ul class="space-y-2">{ for _, r := range rows {
			<UserCard r={r}/>
		} }</ul></main>
	</body></html>
}

// UserCard is the per-row component: more utility classes, a dynamic URL, a
// boolean attribute, and a conditional block.
component UserCard(r data.Row) {
	<li class="rounded border bg-white p-4 shadow-sm"><div class="flex items-center justify-between"><div class="min-w-0"><a class="font-medium text-gray-900 hover:underline" href={r.Href()}>{ r.Name }</a><p class="truncate text-sm text-gray-500">{ r.Email }</p></div><span class="rounded bg-gray-100 px-2 py-1 text-xs font-medium uppercase">{ r.Role }</span></div>{ if !r.Active {
		<p class="mt-2 text-xs text-amber-600" hidden>Inactive account</p>
	} }</li>
}
