//go:build !go.1.18
// +build !go.1.18

package main

import (
	"fmt"
)

// Version defined via link flags, for compabitily with Go < 1.18
var Version string

func showVersion() {
	if Version != "" {
		fmt.Printf("benchcheck version: %s\n", Version)
	} else {
		fmt.Println("benchcheck version: no version info")
	}
}
