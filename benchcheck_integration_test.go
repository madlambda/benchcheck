//go:build integration
// +build integration

package benchcheck_test

import (
	"strings"
	"testing"

	"github.com/madlambda/benchcheck"
	"github.com/madlambda/spells/assert"
)

func TestStatModule(t *testing.T) {
	type rangef struct {
		start float64
		end   float64
	}
	type diff struct {
		name  string
		delta rangef
	}
	type result struct {
		metric string
		diffs  []diff
	}
	type testcase struct {
		name   string
		module string
		oldver string
		newver string
		want   []result
	}

	if testing.Short() {
		t.Skip("Skipping in short mode")
		return
	}

	t.Parallel()

	tcases := []testcase{
		{
			name:   "stat benchcheck",
			module: "github.com/madlambda/benchcheck",
			oldver: "0f9165271a00b54163d3fc4c73d52a13c3747a75",
			newver: "e90da7b50cf0e191004809e415c64319465286d7",
			want: []result{
				{
					metric: "time/op",
					diffs: []diff{
						{
							name:  "Fake",
							delta: rangef{start: -80, end: -60},
						},
					},
				},
			},
		},
		{
			name:   "stat benchcheck reversed versions",
			module: "github.com/madlambda/benchcheck",
			oldver: "e90da7b50cf0e191004809e415c64319465286d7",
			newver: "0f9165271a00b54163d3fc4c73d52a13c3747a75",
			want: []result{
				{
					metric: "time/op",
					diffs: []diff{
						{
							name:  "Fake",
							delta: rangef{start: 100, end: 400},
						},
					},
				},
			},
		},
		{
			name:   "stat benchcheck same version",
			module: "github.com/madlambda/benchcheck",
			oldver: "e90da7b50cf0e191004809e415c64319465286d7",
			newver: "e90da7b50cf0e191004809e415c64319465286d7",
			want: []result{
				{
					metric: "time/op",
					diffs: []diff{
						{
							name:  "Fake",
							delta: rangef{start: -10, end: 10},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcases {
		tcase := tc

		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			got, err := benchcheck.StatModule(tcase.module, tcase.oldver, tcase.newver)
			assertNoError(t, err)

			// We can't check everything on the result since variance
			// is introduced by changes on the environment (this is an e2e test).
			// We can ensure results inside a reasonably broad delta + function names.

			assert.EqualInts(t, len(tcase.want), len(got), "want %v != got %v", tcase, tcase.want, got)

			for i, gotRes := range got {
				wantRes := tcase.want[i]

				t.Logf("got bench result: %v", gotRes)

				assert.EqualStrings(t, wantRes.metric, gotRes.Metric)

				for j, gotDiff := range gotRes.BenchDiffs {
					wantDiff := wantRes.diffs[j]
					gotName := stripProcCountFromBenchName(gotDiff.Name)
					assert.EqualStrings(t, wantDiff.name, gotName)

					if gotDiff.Delta < wantDiff.delta.start {
						t.Fatalf(
							"got delta %.2f < wanted delta start %.2f",
							gotDiff.Delta, wantDiff.delta.start,
						)
					}
					if gotDiff.Delta > wantDiff.delta.end {
						t.Fatalf(
							"got delta %.2f > wanted delta end %.2f",
							gotDiff.Delta, wantDiff.delta.end,
						)
					}
				}
			}
		})
	}
}

func stripProcCountFromBenchName(name string) string {
	// Benchmark names depend on count of CPUs:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.18.3:src/testing/benchmark.go;drc=47f806ce81aac555946144f112b9f8733e2ed871;l=495
	// Here we remove this info so it is easier to test things independent of env.
	lastIndex := strings.LastIndex(name, "-")
	return name[:lastIndex]
}
