package benchcheck

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// GetModule will download a specific version of a module and
// return a directory where you can find the module code.
// It leverages "go get" to do the job, so the returned directory
// should be considered read only (it is part of the Go cache).
//
// A non-nil error is returned if it fails.
func GetModule(name string, version string) (string, error) {
	// Reference: https://golang.org/ref/mod#go-mod-download
	cmd := exec.Command("go", "mod", "download", "-json", fmt.Sprintf("%s@%s", name, version))
	output, _ := cmd.CombinedOutput()
	// TODO: handle errors

	parsedResult := struct {
		Error string // error loading module
		Dir   string // absolute path to cached source root directory
	}{}
	json.Unmarshal(output, &parsedResult)
	return parsedResult.Dir, nil
}
