// LOVM
//
// Love your Virtual Machines. Or lightweight VM manager.
//
//
package main

import (
	"os"

	cli2 "git.stormbase.io/cbednarski/cli"
	"github.com/cbednarski/lovm/cli"
)

func main() {
	if err := cli.Main(); err != nil {
		cli2.ExitWithError(err)
	}
	os.Exit(0)
}
