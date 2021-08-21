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

```
benchcheck cool.go.module v0.0.1 v0.0.2
```

Comparing performance between two versions of a Go module
and failing on time regression:

```
benchcheck cool.go.module v0.0.1 v0.0.2 -time-delta +13.31%
```

Now doing the same but also checking for allocation regression:

```
benchcheck cool.go.module v0.0.1 v0.0.2 -time-delta +10% -alloc-delta +15% -allocs-delta +20%
```

You can also check if your code got faster and use the check to
I don't know..celebrate ? =P

```
benchcheck cool.go.module v0.0.1 v0.0.2 -time-delta -20%
```
