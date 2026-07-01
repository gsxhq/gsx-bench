package bench

import (
	"strings"
	"testing"

	"github.com/gsxhq/gsx-bench/mergemock"
	"github.com/gsxhq/gsx-bench/templr"
	"github.com/gsxhq/gsx-bench/tw"
)

// Custom (Tailwind-style) merger scenario: 20 buttons whose fallthrough class
// conflicts with the component's base classes, resolved by the configured
// mergemock merger. Both engines call the same merger, so it measures the
// class-merge path under an EXPENSIVE merger (the production setup), not the
// cheap default.
var (
	btnLabels = []string{"Save", "Cancel", "Delete", "Edit", "Share", "Copy", "Move", "Pin", "Flag", "Mute", "Lock", "Tag", "Sort", "Hide", "Open", "Send", "Done", "Next", "Back", "Help"}
	btnOver   = "px-8 bg-red-500"
)

func twGSX() render {
	return gsxRender(tw.Buttons(tw.ButtonsProps{Labels: btnLabels, Override: btnOver}))
}
func twTempl() render { return templRender(templr.Buttons(btnLabels, btnOver)) }

func TestButtonsAgree(t *testing.T) {
	g, tp := renderString(twGSX()), renderString(twTempl())
	if canonical(g) != canonical(tp) {
		t.Fatalf("buttons differ:\n gsx:   %s\n templ: %s", g, tp)
	}
	if !strings.Contains(g, "px-8") || strings.Contains(g, "px-4") {
		t.Fatalf("merger did not resolve px-4 -> px-8: %s", g)
	}
}

func BenchmarkButtonsGSXPooled(b *testing.B)   { pooled(b, twGSX()) }
func BenchmarkButtonsTemplPooled(b *testing.B) { pooled(b, twTempl()) }

// TestButtonsMergerCalls pins the fallthrough class merge count. gsx should
// merge each caller-provided class exactly once at the child root, matching
// templ's merger call count for the same 20 buttons.
func TestButtonsMergerCalls(t *testing.T) {
	mergemock.Calls.Store(0)
	_ = renderString(twGSX())
	gsxCalls := mergemock.Calls.Load()

	mergemock.Calls.Store(0)
	_ = renderString(twTempl())
	templCalls := mergemock.Calls.Load()

	t.Logf("merger calls for 20 buttons — gsx: %d, templ: %d", gsxCalls, templCalls)
	if gsxCalls != templCalls {
		t.Fatalf("merger call count mismatch: gsx %d, templ %d", gsxCalls, templCalls)
	}
}
