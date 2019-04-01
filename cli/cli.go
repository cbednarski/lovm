package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/cbednarski/lovm/engine"
	"github.com/cbednarski/lovm/vm"
	"github.com/cbednarski/lovm/vmware"
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

const CommandHelp = `Commands

  lovm clone <source>                   Clone a VM. Start here!
  lovm start                            Start the VM
  lovm stop                             Stop the VM
  lovm restart                          Stop / start the VM
  lovm ssh                              Open an SSH session to the VM
  lovm ip                               Write the VM's IP address to stdout
  lovm mount <host path> <guest path>   Mount a host folder into the VM
  lovm delete                           Delete the VM
  lovm help                             Show help
`

const ProgramHelp = `LOVM

  A minimalist, idempotent command-line tool for managing local virtual machines

`+ CommandHelp +`
Misc

  Copyright: 2019 Chris Bednarski
  License: MIT
  Contact: https://github.com/cbednarski/lovm
`

func Main() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	command, args := ParseArgs(os.Args)

	machine := &vm.VirtualMachine{}

	// TODO Add some error handling if the file exists but we can't read it
	if t, err := vm.VirtualMachineFromFile(pwd); err == nil {
		machine = t
	}

	// TODO generalize this for other virt engines (maybe)
	engine := vmware.New(machine)

	// TODO Implement the rest of the CLI
	switch command {
	case "-h":
		fallthrough
	case "--help":
		fallthrough
	case "help":
		fmt.Print(ProgramHelp)
		return nil
	case "clone":
		if err := ParseClone(args, machine); err != nil {
			return err
		}
		if err := engine.Clone(machine.Source); err != nil {
			return err
		}
	case "start":
		if err := engine.Start(); err != nil {
			return err
		}
		fmt.Printf("machine %q running (%s)\n", machine.Path, machine.Engine)
	case "stop":
		if err := engine.Stop(); err != nil {
			return err
		}
	case "restart":
		if err := engine.Restart(); err != nil {
			return err
		}
	case "ssh":
	case "ip":
	case "mount":
		if err := ParseMounts(args, machine); err != nil {
			return err
		}
		if err := engine.Mount(); err != nil {
			return err
		}
	case "delete":
		if err := engine.Delete(); err != nil {
			return err
		}
	default:
		fmt.Printf("Unknown command %q\n\n", command)
		fmt.Print(CommandHelp)
		return nil
	}

	// If the command ran successfully we'll save and update the machine file.
	// If there was an error earlier we should have aborted already and we'll
	// leave the machine file alone.
	if err := machine.Save(pwd); err != nil {
		return fmt.Errorf("error writing changes to %s: %s", vm.MachineFile, err)
	}

	return nil
}
