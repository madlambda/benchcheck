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
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("error fetching module: %v : output: %v", err, output)
	}

	parsedResult := struct {
		Dir string // absolute path to cached source root directory
	}{}

	err = json.Unmarshal(output, &parsedResult)
	if err != nil {
		return "", fmt.Errorf("error parsing %q : %v", string(output), err)
	}
	return parsedResult.Dir, nil
}
