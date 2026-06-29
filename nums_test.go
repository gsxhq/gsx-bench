package bench

import (
	"testing"

	"github.com/gsxhq/gsx-bench/gsxr"
	"github.com/gsxhq/gsx-bench/templr"
)

func statsGSX() render   { return gsxRender(gsxr.Stats(gsxr.StatsProps{Rows: rows})) }
func statsTempl() render { return templRender(templr.Stats(rows)) }

func TestStatsAgree(t *testing.T) {
	if canonical(renderString(statsGSX())) != canonical(renderString(statsTempl())) {
		t.Fatalf("stats differ:\n gsx:   %s\n templ: %s", renderString(statsGSX()), renderString(statsTempl()))
	}
}

func BenchmarkStatsGSXPooled(b *testing.B)   { pooled(b, statsGSX()) }
func BenchmarkStatsTemplPooled(b *testing.B) { pooled(b, statsTempl()) }
