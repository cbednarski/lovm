package cli

import (
	"errors"

	"github.com/cbednarski/lovm/engine"
	"github.com/cbednarski/lovm/vm"
)

func ParseArgs(input []string) (command string, args []string) {
	// We're getting os.Args verbatim, so input[0] is always the name of the
	// program. input[1] is the command, and everything else are arguments. If
	// len(input) is less than 2 the user did not type a command so we'll just
	// return here.
	if len(input) > 1 {
		command = input[1]
	}
	if len(input) > 2 {
		args = input[2:]
	}

	return
}

func ParseClone(args []string, machine *vm.VirtualMachine) error {
	// We accept 0 or 1 arguments because we can use the clone source already
	// configured in machine.lovm (if it exists). If machine.Source is null
	// we'll complain.
	switch len(args) {
	case 0:
		if machine.Source == "" {
			return errors.New("clone source must be specified, e.g. /path/to/some.vmx")
		}
	case 1:
		machine.Engine = engine.Identify(args[0])
	default:
		return errors.New("too many arguments")
	}

	return nil
}

func ParseMounts(args []string, machine *vm.VirtualMachine) error {
	if len(args) != 2 {
		return errors.New("expected args <host path to mount> <target path in guest>")
	}

	machine.Mounts[args[0]] = args[1]

	return nil
}
