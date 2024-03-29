package benchcheck

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/perf/benchstat"
)

// CheckerFmt represents the expected string format of a checker.
const CheckerFmt = "<metric>=(+|-)<number>%"

// Module represents a Go module.
type Module struct {
	path string
}

// StatResult is the full result showing performance
// differences between two benchmark runs (set of benchmark functions)
// for a specific metric, like time/op or speed.
type StatResult struct {
	// Metric is the name of metric
	Metric string
	// BenchDiffs has the performance diff of all function for a given metric.
	BenchDiffs []BenchDiff
}

// BenchResults represents a single Go benchmark run. Each
// string is the result of a single Benchmark like this:
// - "BenchmarkName  	 50	  31735022 ns/op	  61.15 MB/s"
type BenchResults []string

// BenchDiff is the result showing performance differences
// for a single benchmark function.
type BenchDiff struct {
	// Name of the benchmark function
	Name string
	// Old is the performance summary of the old benchmark.
	Old string
	// New is the performance summary of the new benchmark.
	New string
	// Delta between the old and new performance summaries.
	Delta float64
}

// Checker performs checks on StatResult.
type Checker struct {
	metric    string
	threshold float64
	repr      string
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

// String returns the string representation of the checker.
func (c Checker) String() string {
	return c.repr
}

// Do performs the check on the given StatResult. Returns true
// if it passed the check, false otherwise.
func (c Checker) Do(stat StatResult) bool {
	if c.metric != stat.Metric {
		return true
	}

	for _, bench := range stat.BenchDiffs {
		if c.threshold >= 0.0 && bench.Delta > c.threshold {
			return false
		}
		if c.threshold < 0.0 && bench.Delta < c.threshold {
			return false
		}
	}
	return true
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
func (b BenchDiff) String() string {
	return fmt.Sprintf(
		"%s: old %s: new %s: delta: %.2f%%",
		b.Name, b.Old, b.New, b.Delta,
	)
}

// Add will add a new bench result. If the string doesn't represent
// a benchmark result it will be ignored.
func (b *BenchResults) Add(res string) {
	if !strings.HasPrefix(res, "Benchmark") {
		return
	}
	*b = append(*b, res)
}

// GetModule will download a specific version of a module and
// return a directory where you can find the module code.
// It uses "go get" to do the job, so the returned directory
// should be considered read only (it is part of the Go cache).
// The returned path is an absolute path.
//
// Any errors running "go" can be inspected in detail by
// checking if the returned error is a *CmdError.
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

// RunBench will run all benchmarks present at the given module
// return the benchmark results.
//
// This function relies on running the "go" command to run benchmarks.
//
// Any errors running "go" can be inspected in detail by
// checking if the returned is a *CmdError.
func RunBench(mod Module) (BenchResults, error) {
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
	results := BenchResults{}
	for _, b := range benchOut {
		results.Add(b)
	}
	return results, nil
}

// Stat compares two benchmark results providing a set of stats results.
func Stat(oldres BenchResults, newres BenchResults) ([]StatResult, error) {
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

// StatModule will:
//
// - Download the specific versions of the given module.
// - Run benchmarks on each of them.
// - Compare old vs new version benchmarks and return a stat results.
//
// This function relies on running the "go" command to run benchmarks.
//
// Any errors running "go" can be inspected in detail by
// checking if the returned error is a CmdError.
func StatModule(name string, oldversion, newversion string) ([]StatResult, error) {
	oldresults, err := benchModule(name, oldversion)
	if err != nil {
		return nil, fmt.Errorf("running bench for old module: %v", err)
	}

	newresults, err := benchModule(name, newversion)
	if err != nil {
		return nil, fmt.Errorf("running bench for new module: %v", err)
	}

	return Stat(oldresults, newresults)
}

// ParseChecker will parse the given string into a Check.
func ParseChecker(val string) (Checker, error) {
	parsed := strings.Split(val, "=")
	metric := parsed[0]
	if len(parsed) != 2 {
		return Checker{}, fmt.Errorf("checker on wrong format, expect: %q", CheckerFmt)
	}
	threshold, err := strconv.ParseFloat(strings.TrimSuffix(parsed[1], "%"), 64)
	if err != nil {
		return Checker{}, fmt.Errorf("parsing checker %q threshold", err)
	}
	return Checker{
		repr:      val,
		metric:    metric,
		threshold: threshold,
	}, nil
}

func newStatResults(tables []*benchstat.Table) []StatResult {
	res := make([]StatResult, len(tables))

	for i, table := range tables {
		res[i] = StatResult{
			Metric:     table.Metric,
			BenchDiffs: newBenchResults(table.Rows),
		}
	}

	return res
}

func newBenchResults(rows []*benchstat.Row) []BenchDiff {
	res := make([]BenchDiff, len(rows))

	for i, row := range rows {
		if len(row.Metrics) != 2 {
			// Since we always compare 2 sets of benchmarks there should be
			// 2 metrics per row always.
			panic(fmt.Errorf("should always have 2 metrics, instead got: %d", len(row.Metrics)))
		}

		res[i] = BenchDiff{
			Name:  row.Benchmark,
			Old:   row.Metrics[0].Format(row.Scaler),
			New:   row.Metrics[1].Format(row.Scaler),
			Delta: row.PctDelta,
		}
	}

	return res
}

func resultsReader(res BenchResults) io.Reader {
	return strings.NewReader(strings.Join(res, "\n"))
}

func benchModule(name string, version string) (BenchResults, error) {
	mod, err := GetModule(name, version)
	if err != nil {
		return nil, err
	}
	// benchstat requires multiple runs of the same benchmarks
	// so it can assess statistically for abnormalities, etc.
	const benchruns = 5

	results := BenchResults{}

	for i := 0; i < benchruns; i++ {
		res, err := RunBench(mod)
		if err != nil {
			return nil, err
		}
		results = append(results, res...)
	}

	return results, nil
}
