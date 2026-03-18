//go:build !mage
// +build !mage

package main

import (
	"flat/cmd"
)

func main() {
	cmd.Execute()
}
