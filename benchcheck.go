package benchcheck

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"golang.org/x/perf/benchstat"
)

// Module represents a Go module.
type Module struct {
	path string
}

// StatResult is the full result showing performance
// differences between two benchmark runs (set of benchmark functions)
// for a specific metric, like time/op or speed.
type StatResult struct {
	// Name of metric
	Metric string
	// List of results for a specific metric
	BenchResults []BenchResult
}

// BenchResult is the result showing performance differences
// for a single benchmark function.
type BenchResult struct {
	// Name of the benchmark function
	Name string
	// Old is the performance summary of the old benchmark.
	Old string
	// New is the performance summary of the new benchmark.
	New string
	// Delta between the old and new performance summaries.
	Delta float64
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

// String provides the string representation of a bench result
func (br BenchResult) String() string {
	return fmt.Sprintf(
		"%s: old %s: new %s: delta: %.2f%%",
		br.Name, br.Old, br.New, br.Delta,
	)
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
// of a benchmark from Go test.
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

// Stat compares two benchmark results providing a set of results.
// oldres and newres must be multiple strings where each line follows
// Go's benchmark output format.
func Stat(oldres []string, newres []string) ([]StatResult, error) {
	// We are using benchstat defaults:
	//	- https://cs.opensource.google/go/x/perf/+/master:cmd/benchstat/main.go;l=117
	const (
		alpha   = 0.05
		geomean = false
	)
	c := &benchstat.Collection{
		Alpha:      alpha,
		AddGeoMean: geomean,
		DeltaTest:  benchstat.UTest,
	}
	if err := c.AddFile("old", resultsReader(oldres)); err != nil {
		return nil, fmt.Errorf("parsing old results: %v", err)
	}
	if err := c.AddFile("new", resultsReader(newres)); err != nil {
		return nil, fmt.Errorf("parsing new results: %v", err)
	}
	return newStatResults(c.Tables()), nil
}

func newStatResults(tables []*benchstat.Table) []StatResult {
	res := make([]StatResult, len(tables))

	for i, table := range tables {
		res[i] = StatResult{
			Metric:       table.Metric,
			BenchResults: newBenchResults(table.Rows),
		}
	}

	return res
}

func newBenchResults(rows []*benchstat.Row) []BenchResult {
	res := make([]BenchResult, len(rows))

	for i, row := range rows {
		if len(row.Metrics) != 2 {
			// Since we always compare 2 sets of benchmarks there should be
			// 2 metrics per row always.
			panic(fmt.Errorf("should always have 2 metrics, instead got: %d", len(row.Metrics)))
		}

		res[i] = BenchResult{
			Name:  row.Benchmark,
			Old:   row.Metrics[0].Format(row.Scaler),
			New:   row.Metrics[1].Format(row.Scaler),
			Delta: row.PctDelta,
		}
	}

	return res
}

func resultsReader(res []string) io.Reader {
	return strings.NewReader(strings.Join(res, "\n"))
}
