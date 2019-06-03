// LOVM
//
// Love your Virtual Machines. Or lightweight VM manager.
//
//
package main

import (
	"os"

	"github.com/cbednarski/cli"
	"github.com/cbednarski/lovm/commands"
)

func main() {
	if err := commands.Main(); err != nil {
		cli.ExitWithError(err)
	}
	os.Exit(0)
}
