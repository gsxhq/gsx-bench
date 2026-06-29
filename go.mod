module github.com/gsxhq/gsx-bench

go 1.26.1

// gsx is benchmarked against its local working tree, not a published release,
// so the numbers track main as it evolves. templ, by contrast, is pulled from
// its published module (origin main), never a local experimental checkout.
replace github.com/gsxhq/gsx => ../gsx

require (
	github.com/a-h/templ v0.3.1020
	github.com/gsxhq/gsx v0.0.0-20260628200854-1920733db19d
)
