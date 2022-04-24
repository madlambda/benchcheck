package benchcheck

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Module represents a Go module.
type Module struct {
	path string
}

// CmdError represents an error running a specific command.
type CmdError struct {
	Cmd    *exec.Cmd
	Err    error
	Output string
}

// Error returns the string representation of the error.
func (c *CmdError) Error() string {
	return fmt.Sprintf(
		"fail to run: %v on dir %s: %v",
		c.Cmd,
		c.Cmd.Dir,
		c.Err,
	)
}

// Path is the absolute path of the module on the filesystem.
func (m Module) Path() string {
	return m.path
}

// String provides the string representation of the module.
func (m Module) String() string {
	return fmt.Sprintf("go module at %q", m.path)
}

// GetModule will download a specific version of a module and
// return a directory where you can find the module code.
// It uses "go get" to do the job, so the returned directory
// should be considered read only (it is part of the Go cache).
// The returned path is an absolute path.
//
// Any errors running "go" can be inspected in detail by
// checking if the returned is a CmdError.
func GetModule(name string, version string) (Module, error) {
	// Reference: https://golang.org/ref/mod#go-mod-download
	cmd := exec.Command("go", "mod", "download", "-json", fmt.Sprintf("%s@%s", name, version))
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Module{}, &CmdError{
			Cmd:    cmd,
			Err:    err,
			Output: string(output),
		}
	}

	parsedResult := struct {
		Dir string // absolute path to cached source root directory
	}{}

	err = json.Unmarshal(output, &parsedResult)
	if err != nil {
		return Module{}, fmt.Errorf("error parsing %q : %v", string(output), err)
	}
	return Module{path: parsedResult.Dir}, nil
}

// RunBench will run all benchmarks present at the given module.
// It returns an slice of strings where each string is the result
// of a benchmark.
//
// This function relies on running the "go" command to run benchmarks.
// Any errors running "go" can be inspected in detail by
// checking if the returned is a CmdError.
func RunBench(mod Module) ([]string, error) {
	cmd := exec.Command("go", "test", "-bench=.", "./...")
	cmd.Dir = mod.Path()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, &CmdError{
			Cmd:    cmd,
			Err:    err,
			Output: string(out),
		}
	}
	benchOut := strings.Split(string(out), "\n")
	results := []string{}
	for _, b := range benchOut {
		if strings.HasPrefix(b, "Benchmark") {
			results = append(results, b)
		}
	}
	return results, nil
}
