// Package mergemock is a stand-in for a real Tailwind-aware class merger
// (e.g. tailwind-merge-go) for benchmarking the configurable-ClassMerger path
// WITHOUT taking that dependency. It is representative, not correct: it resolves
// conflicts by utility "group" (the prefix before the last '-'), last-wins, and
// caches by input like a real merger does. Both the gsx and templ button
// scenarios call it, so the comparison stays apples-to-apples.
package mergemock

import (
	"strings"
	"sync"
	"sync/atomic"
)

var cache sync.Map // joined-input key -> merged result

// Calls counts every Merge invocation (cache hit or miss). Tests use it to show
// how many times each engine runs the merger — exposing gsx's redundant
// double-merge of fallthrough classes independent of merger cost.
var Calls atomic.Int64

// Merge takes the raw on-class strings (gsx's ClassMerger seam) and returns the
// conflict-resolved class string.
func Merge(classes []string) string {
	Calls.Add(1)
	key := strings.Join(classes, "\x00")
	if v, ok := cache.Load(key); ok {
		return v.(string)
	}
	var toks []string
	for _, c := range classes {
		toks = append(toks, strings.Fields(c)...)
	}
	lastByGroup := make(map[string]int, len(toks))
	for i, t := range toks {
		lastByGroup[group(t)] = i
	}
	out := make([]string, 0, len(toks))
	for i, t := range toks {
		if lastByGroup[group(t)] == i {
			out = append(out, t)
		}
	}
	res := strings.Join(out, " ")
	cache.Store(key, res)
	return res
}

// group returns the conflict group of a utility class: everything up to and
// including the last '-' (so px-4 and px-8 share group "px-", and a no-dash
// class like "rounded" is its own group).
func group(t string) string {
	if i := strings.LastIndexByte(t, '-'); i >= 0 {
		return t[:i+1]
	}
	return t
}
