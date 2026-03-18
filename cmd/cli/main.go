//go:build !mage
// +build !mage

package main

import (
	"flat/cmd"
	"os"
)

func main() {
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
