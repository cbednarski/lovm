package cli

import (
	"errors"

	"github.com/cbednarski/lovm/lovm"
)

func ParseArgs(input []string) (command string, args []string) {
	if len(input) < 2 {
		return
	}

	command = input[1]
	args = input[2:]
	return
}

func ParseClone(args []string, machine *lovm.VirtualMachine) error {
	if len(args) == 0 {
		return errors.New("must specify clone source, e.g. /path/to/blah.vmx")
	}
	if len(args) > 1 {
		return errors.New("unexpected arguments")
	}

	machine.Engine = lovm.Identify(args[0])
	machine.Source = args[0]

	return nil
}

func ParseMounts(args []string, machine *lovm.VirtualMachine) error {
	if len(args) != 2 {
		return errors.New("expected args <host path to mount> <target path in guest>")
	}

	machine.Mounts[args[0]] = args[1]

	return nil
}
