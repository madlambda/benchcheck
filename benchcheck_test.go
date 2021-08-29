package benchcheck_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madlambda/benchcheck"
)

func TestGetModule(t *testing.T) {
	// Not extremely happy with the way the tests are coupled to changes
	// on jtoh, like new releases would break latest logic here.
	// But not testing this integrated with actual repos would be almost useless.
	tests := []struct {
		desc          string
		moduleName    string
		moduleVersion string
		wantErr       bool
		wantVersion   string
	}{
		{
			desc:          "ValidVersion",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "v0.1.0",
		},
		{
			desc:          "UsingLatest",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "latest",
			wantVersion:   "v0.1.0",
		},
		{
			desc:          "UsingCommitSha",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "5cd825858d7dcc41c3b453ed10ecf0983b139243",
			wantVersion:   "v0.1.1-0.20210731193031-5cd825858d7d",
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
			modDir, err := benchcheck.GetModule(test.moduleName, test.moduleVersion)

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

			fileinfo, err := os.Stat(modDir)

			if err != nil {
				t.Fatalf("os.Stat(%q): unexpected error : %v", modDir, err)
			}

			if !fileinfo.IsDir() {
				t.Fatalf("want %q to be a dir, details : %v", modDir, fileinfo)
			}

			if test.wantVersion == "" {
				test.wantVersion = test.moduleVersion
			}

			gotModuleVersion := getModuleVersion(t, modDir)
			if gotModuleVersion != test.wantVersion {
				t.Fatalf(
					"got version %q from mod dir %q, wanted %q",
					gotModuleVersion,
					modDir,
					test.wantVersion,
				)
			}
		})
	}
}

func getModuleVersion(t *testing.T, modDir string) string {
	// This function abuses private details about the modDir,
	// more specifically, the fact it is a Go cache dir.
	// But it is at least decoupled from specific version systems.
	// We couple on whatever interfaces Go provides.
	modNameVersion := filepath.Base(modDir)
	parsed := strings.Split(modNameVersion, "@")
	if len(parsed) <= 1 {
		t.Fatalf("module cache dir supposed to be on the form 'module@version' got: %q", modDir)
	}
	return strings.Join(parsed[1:], "")
}
