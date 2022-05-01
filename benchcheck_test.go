package benchcheck_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	assertNoError(t, err, "benchcheck.GetModule(%q, %q)", module, modversion)

	res, err := benchcheck.RunBench(mod)
	assertNoError(t, err, "benchcheck.RunBench(%v)", mod)

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
		module     = "github.com/madlambda/benchchec"
		modversion = "f15923bf230cc7331ad869fcdaac35172f8b7f38"
	)
	mod, err := benchcheck.GetModule(module, modversion)
	assertNoError(t, err, "benchcheck.GetModule(%q, %q)", module, modversion)

	res, err := benchcheck.RunBench(mod)
	assertNoError(t, err, "benchcheck.RunBench(%v)", mod)

	assert.EqualInts(t, 0, len(res), "want no results, got: %v", res)
}

func assertNoError(t *testing.T, err error, details ...interface{}) {
	t.Helper()

	if err == nil {
		return
	}

	msg := ""

	if len(details) > 0 {
		msg = fmt.Sprintf(details[0].(string), details[1:]...) + ":"
	}

	msg += err.Error()

	var cmderr *benchcheck.CmdError
	if errors.As(err, &cmderr) {
		msg += "\ncmd stdout + stderr:\n" + cmderr.Output
	}

	t.Fatal(msg)
}
