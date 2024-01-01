# benchcheck

**benchcheck** runs benchmarks on two different versions of a Go module and
compares statistics across them, failing on performance regressions.

This tool aims at being Go specific and wildly simple, leveraging standard Go
benchmarks + [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat),
composing them to provide just a few extra features:

* Given two module versions of a package, it automatically benchmarks and compares them
* Given a threshold on the delta of performance loss, the check fails (useful for CI's)

And that is it, no more, no less.

# Dependencies

* git
* go

# Install

Just run:

```sh
go install github.com/madlambda/benchcheck/cmd/benchcheck@latest
```

# Usage

For details just run:

```sh
benchcheck -help
```

By default all benchmarks are run with -benchmem, so you also
have information on memory allocations.

Comparing performance between two versions of a Go module
and just showing results on output (no check performed):

```sh
benchcheck -mod cool.go.module -old v0.0.1 -new v0.0.2
```

Comparing performance between two versions of a Go module
and failing on time regression:

```sh
benchcheck -mod cool.go.module -old v0.0.1 -new v0.0.2 -check time/op=13.31%
```

You can also check if your code got faster and use the check to
I don't know... Celebrate ? =P

```sh
benchcheck -mod cool.go.module -old v0.0.1 -new v0.0.2 -check time/op=-13.31%
```

Now lets say you want to customize how the benchmarks are run, just add the command that you wish
to be executed to run the benchmarks like this:

```sh
benchcheck -mod cool.go.module -old v0.0.1 -new v0.0.2 -- go test -bench=BenchmarkSpecific ./specific/pkg
```

It can be any command that will generate benchmark results with the same formatting as `go test` benchmarks.
