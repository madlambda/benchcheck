package benchcheck_test

import (
	"os"
	"testing"

	"github.com/madlambda/benchcheck"
)

func TestGetModule(t *testing.T) {
	tests := []struct {
		desc          string
		moduleName    string
		moduleVersion string
		wantErr       bool
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
		},
		{
			desc:          "UsingCommitSha",
			moduleName:    "github.com/madlambda/jtoh",
			moduleVersion: "5cd825858d7dcc41c3b453ed10ecf0983b139243",
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
		})
	}
}
