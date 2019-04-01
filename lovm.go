// LOVM
//
// Love your Virtual Machines. Or lightweight VM manager.
//
//
package main

import (
	"fmt"
	"os"

	"github.com/cbednarski/lovm/cli"
	"github.com/cbednarski/lovm/vm"
	"github.com/cbednarski/lovm/vmware"
)

const commandText = `Commands

  lovm clone <source>                   Clone a VM. Start here!
  lovm start                            Start the VM
  lovm stop                             Stop the VM
  lovm restart                          Stop / start the VM
  lovm ssh                              Open an SSH session to the VM
  lovm ip                               Write the VM's IP address to stdout
  lovm mount <host path> <guest path>   Mount a host folder into the VM
  lovm delete                           Delete the VM
`

const helpText = `LOVM

  A minimalist command-line utility for managing local virtual machines

`+commandText+`
Misc

  Copyright: 2019 Chris Bednarski
  License: MIT
  Contact: https://github.com/cbednarski/lovm
`

func wrappedMain() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	command, args := cli.ParseArgs(os.Args)

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
		fmt.Print(helpText)
		return nil
	case "clone":
		if err := cli.ParseClone(args, machine); err != nil {
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
	case "ssh":
	case "ip":
	case "mount":
		if err := cli.ParseMounts(args, machine); err != nil {
			return err
		}
	case "delete":
		if err := engine.Delete(); err != nil {
			return err
		}
	default:
		fmt.Printf("Unknown command %q\n\n", command)
		fmt.Print(commandText)
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

func main() {
	err := wrappedMain()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	os.Exit(0)
}
