package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/madlambda/benchcheck"
)

type checkList []benchcheck.Checker

func (c *checkList) String() string {
	if c == nil {
		return ""
	}
	strs := make([]string, len(*c))
	for _, check := range *c {
		strs = append(strs, check.String())
	}
	return strings.Join(strs, ",")
}

func (c *checkList) Set(val string) error {
	check, err := benchcheck.ParseChecker(val)
	if err != nil {
		return err
	}
	*c = append(*c, check)
	return nil
}

func main() {
	version := flag.Bool("version", false, "show version")
	mod := flag.String("mod", "", "module to be bench checked")
	oldRev := flag.String("old", "", "the old revision to compare")
	newRev := flag.String("new", "", "the new revision to compare")

	checks := checkList{}
	flag.Var(&checks, "check", fmt.Sprintf(
		"check to be performed, defined in the form: %s. Eg: time/op=10%%",
		benchcheck.CheckerFmt))

	flag.Parse()

	if *version {
		showVersion()
		return
	}

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	if *mod == "" {
		log.Fatal("-mod is obligatory")
	}
	if *oldRev == "" {
		log.Fatal("-old is obligatory")
	}
	if *newRev == "" {
		log.Fatal("-new is obligatory")
	}

	runBench := benchcheck.DefaultRunBench
	var customizedBenchCmd []string

	for i, v := range os.Args {
		if v == "--" {
			customizedBenchCmd = os.Args[i+1:]
		}
	}

	if len(customizedBenchCmd) > 0 {
		runBench = benchcheck.NewBenchRunner(customizedBenchCmd[0], customizedBenchCmd[1:]...)
	}

	results, err := benchcheck.StatModule(runBench, *mod, *oldRev, *newRev)
	if err != nil {
		var cmderr *benchcheck.CmdError
		if errors.As(err, &cmderr) {
			fmt.Fprintf(os.Stderr, "failed to run: %s", cmderr.Cmd)
			fmt.Fprintf(os.Stderr, "error: %s", cmderr.Err)
			fmt.Fprintf(os.Stderr, "cmd output: %s", cmderr.Output)
		} else {
			log.Fatal(err)
		}
	}

	for _, result := range results {
		fmt.Printf("metric: %s\n", result.Metric)
		for _, diff := range result.BenchDiffs {
			fmt.Println(diff)
		}
		for _, check := range checks {
			if !check.Do(result) {
				fmt.Printf("check failed: %s", check)
			}
		}
	}
}
