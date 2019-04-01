// LOVM
//
// Love your Virtual Machines. Or lightweight VM manager.
//
//
package main

import (
	"os"

	"github.com/cbednarski/lovm/cli"
)

func main() {
	err := cli.Main()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	os.Exit(0)
}
