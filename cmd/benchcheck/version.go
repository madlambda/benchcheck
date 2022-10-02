//go:build 1.18

package main

import (
	"fmt"
	"runtime/debug"
)

func showVersion() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				fmt.Printf("go version: %s\n", info.GoVersion)
				fmt.Printf("benchcheck version: %s\n", setting.Value)
				return
			}
		}
	}
	fmt.Println("version: no version info")
}
