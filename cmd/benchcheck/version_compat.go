//go:build !go.1.18

package main

import (
	"fmt"
	"runtime/debug"
)

// Version defined via link flags, for compabitily with Go < 1.18
var Version string

func showVersion() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		fmt.Printf("go version: %s\n", info.GoVersion)
	}
	if Version != "" {
		fmt.Printf("benchcheck version: %s\n", Version)
	}
}
