// LOVM
//
// Love your Virtual Machines. Or lightweight VM manager.
//
//
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cbednarski/lovm/cli"
	"github.com/cbednarski/lovm/lovm"
	"github.com/cbednarski/lovm/vmware"
)

const helpText = `LOVM

  A minimalist command-line utility for managing local virtual machines

Commands

  lovm clone <source>                   Clone a VM. Start here!
  lovm start                            Start the VM
  lovm stop                             Stop the VM
  lovm restart                          Restart, or poweroff/poweron the VM
  lovm ssh                              Open an SSH session to the VM
  lovm ip                               Write the VM's IP address to stdout
  lovm mount <host path> <guest path>   Mount a host folder into the VM

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

	machine := &lovm.VirtualMachine{}

	// TODO Add some error handling if the file exists but we can't read it
	if t, err := lovm.VirtualMachineFromFile(pwd); err == nil {
		machine = t
	}

	// TODO generalize this for other virt engines (maybe)
	engine := vmware.New(machine)

	// TODO Implement the rest of the CLI
	switch command {
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
		log.Printf("machine %q running (%s)\n", machine.Path, machine.Engine)
	case "stop":
	case "restart":
	case "ssh":
	case "ip":
	case "mount":
		if err := cli.ParseMounts(args, machine); err != nil {
			return err
		}
	default:
		fmt.Print(helpText)
		return nil
	}

	// If the command ran successfully we'll save and update the lovm file. If
	// there was an error we'll abort before we get here and leave it alone.
	if err := machine.Save(pwd); err != nil {
		return fmt.Errorf("error writing changes to %s: %s", lovm.LovmFile, err)
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
