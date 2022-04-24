package benchcheck_test

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/madlambda/benchcheck"
	"github.com/madlambda/spells/assert"
)

func TestGetModule(t *testing.T) {
	t.Parallel()

	type ModuleInfo struct {
		Name    string
		Version string
	}

	getModuleInfo := func(t *testing.T, modDir string) ModuleInfo {
		modNameVersion := filepath.Base(modDir)
		parsed := strings.Split(modNameVersion, "@")
		if len(parsed) <= 1 {
			t.Fatalf("module cache dir supposed to be on the form 'module@version' got: %q", modDir)
		}
		return ModuleInfo{
			Name:    parsed[0],
			Version: strings.Join(parsed[1:], ""),
		}
	}

	tests := []struct {
		desc          string
		moduleName    string
		moduleVersion string
		wantErr       bool
		wantModInfo   ModuleInfo
	}{
		{
			desc:          "ValidVersion",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "v0.1.0",
			wantModInfo: ModuleInfo{
				Name:    "jtoh",
				Version: "v0.1.0",
			},
		},
		{
			desc:          "UsingLatest",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "latest",
			wantModInfo: ModuleInfo{
				Name:    "jtoh",
				Version: "v0.1.0",
			},
		},
		{
			desc:          "UsingCommitSha",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "5cd825858d7dcc41c3b453ed10ecf0983b139243",
			wantModInfo: ModuleInfo{
				Name:    "jtoh",
				Version: "v0.1.1-0.20210731193031-5cd825858d7d",
			},
		},
		{
			desc:          "InvalidVersion",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "StoNkS",
			wantErr:       true,
		},
		{
			desc:          "InvalidModule",
			moduleName:    "git.duh/suchwrong/muchfail",
			moduleVersion: "latest",
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mod, err := benchcheck.GetModule(test.moduleName, test.moduleVersion)

			if test.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf(
					"benchcheck.GetModule(%q, %q): unexpected error : %v",
					test.moduleName,
					test.moduleVersion,
					err,
				)
			}

			fileinfo, err := os.Stat(mod.Path())

			if err != nil {
				t.Fatalf("os.Stat(%q): unexpected error : %v", mod.Path(), err)
			}

			if !fileinfo.IsDir() {
				t.Fatalf("want %q to be a dir, details : %v", mod.Path(), fileinfo)
			}

			gotModInfo := getModuleInfo(t, mod.Path())
			if gotModInfo != test.wantModInfo {
				t.Fatalf(
					"got module %v from mod dir %q, wanted %v",
					gotModInfo,
					mod.Path(),
					test.wantModInfo,
				)
			}
		})
	}
}

func TestBenchModule(t *testing.T) {
	t.Parallel()

	const (
		module     = "github.com/madlambda/benchcheck"
		modversion = "73348d58a038746fd4f92dd1e77344a58a4f8505"
	)
	mod, err := benchcheck.GetModule(module, modversion)
	assert.NoError(t, err, "benchcheck.GetModule(%q, %q)", module, modversion)

	res, err := benchcheck.RunBench(mod)
	assert.NoError(t, err, "benchcheck.RunBench(%v)", mod)

	assert.EqualInts(t, 1, len(res), "want single result, got: %v", res)
	if !strings.HasPrefix(res[0], "BenchmarkFake") {
		t.Fatalf("bench result has wrong prefix: %s", res[0])
	}
	if !strings.Contains(res[0], "ns/op") {
		t.Fatalf("bench result should contain time info: %s", res[0])
	}
}

func TestBenchModuleNoBenchmarks(t *testing.T) {
	t.Parallel()

	const (
		module     = "github.com/madlambda/benchcheck"
		modversion = "f15923bf230cc7331ad869fcdaac35172f8b7f38"
	)
	mod, err := benchcheck.GetModule(module, modversion)
	assert.NoError(t, err, "benchcheck.GetModule(%q, %q)", module, modversion)

	res, err := benchcheck.RunBench(mod)
	assert.NoError(t, err, "benchcheck.RunBench(%v)", mod)

	assert.EqualInts(t, 0, len(res), "want no results, got: %v", res)
}

func TestStatBenchmarkResults(t *testing.T) {
	type testcase struct {
		name   string
		oldres []string
		newres []string
		want   []benchcheck.StatResult
	}

	tcases := []testcase{
		{
			name: "benchstat basic example",
			oldres: []string{
				"BenchmarkGobEncode   	100	  13552735 ns/op	  56.63 MB/s",
				"BenchmarkJSONEncode  	 50	  32395067 ns/op	  59.90 MB/s",
				"BenchmarkGobEncode   	100	  13553943 ns/op	  56.63 MB/s",
				"BenchmarkJSONEncode  	 50	  32334214 ns/op	  60.01 MB/s",
				"BenchmarkGobEncode   	100	  13606356 ns/op	  56.41 MB/s",
				"BenchmarkJSONEncode  	 50	  31992891 ns/op	  60.65 MB/s",
				"BenchmarkGobEncode   	100	  13683198 ns/op	  56.09 MB/s",
				"BenchmarkJSONEncode  	 50	  31735022 ns/op	  61.15 MB/s",
			},
			newres: []string{
				"BenchmarkGobEncode   	 100	  11773189 ns/op	  65.19 MB/s",
				"BenchmarkJSONEncode  	  50	  32036529 ns/op	  60.57 MB/s",
				"BenchmarkGobEncode   	 100	  11942588 ns/op	  64.27 MB/s",
				"BenchmarkJSONEncode  	  50	  32156552 ns/op	  60.34 MB/s",
				"BenchmarkGobEncode   	 100	  11786159 ns/op	  65.12 MB/s",
				"BenchmarkJSONEncode  	  50	  31288355 ns/op	  62.02 MB/s",
				"BenchmarkGobEncode   	 100	  11628583 ns/op	  66.00 MB/s",
				"BenchmarkJSONEncode  	  50	  31559706 ns/op	  61.49 MB/s",
				"BenchmarkGobEncode   	 100	  11815924 ns/op	  64.96 MB/s",
				"BenchmarkJSONEncode  	  50	  31765634 ns/op	  61.09 MB/s",
			},
			want: []benchcheck.StatResult{
				{
					Metric: "time/op",
					BenchResults: []benchcheck.BenchResult{
						{Name: "GobEncode", Delta: -13.3},
						{Name: "JSONEncode", Delta: 0.0},
					},
				},
				{
					Metric: "speed",
					BenchResults: []benchcheck.BenchResult{
						{Name: "GobEncode", Delta: 15.35},
						{Name: "JSONEncode", Delta: 0.0},
					},
				},
			},
		},
	}

	cmpfloats := cmp.Comparer(func(x, y float64) bool {
		const ε = 0.01
		return math.Abs(x-y) < ε && math.Abs(y-x) < ε
	})

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := benchcheck.Stat(tcase.oldres, tcase.newres)
			assert.NoError(t, err)

			if diff := cmp.Diff(got, tcase.want, cmpfloats); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
