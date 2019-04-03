package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine"
	"github.com/cbednarski/lovm/engine/vmware"
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

func ParseClone(args []string, config *core.MachineConfig) (string, error) {
	// We accept 0 or 1 arguments because we can use the clone source already
	// configured in machine.lovm (if it exists). If machine.Source is null
	// we'll complain.
	source := ""
	switch len(args) {
	case 0:
		if config.Source == "" {
			return source, errors.New("clone source must be specified, e.g. /path/to/some.vmx")
		}
	case 1:
		source = args[0]
		config.Engine = engine.Identify(args[0])
	default:
		return source, errors.New("too many arguments")
	}

	return source, nil
}

func ParseMounts(args []string, machine *core.MachineConfig) error {
	if len(args) != 2 {
		return errors.New("expected args <host path to mount> <target path in guest>")
	}

	machine.Mounts[args[0]] = args[1]

	return nil
}

func SSH(args []string, machine core.VirtualizationEngine) error {
	ip, err := machine.IP()
	if err != nil {
		return err
	}

	// Any additional arguments (-i, -l, etc.) may be passed through to the
	// underlying ssh command, while the IP is filled in automatically
	args = append(args, ip.String())

	command := exec.Command("ssh", args...)

	// Pass through stdin, stdout, and stderr to the child process
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Fork the child process
	if err := command.Start(); err != nil {
		return err
	}

	// Wait for it to complete
	if err := command.Wait(); err != nil {
		return err
	}

	return nil
}

// Help allows you to return a nil error after showing help text
func Help(text string) error {
	fmt.Print(text)
	return nil
}

func Main() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	command, args := ParseArgs(os.Args)

	config := &core.MachineConfig{}

	// TODO Add some error handling if the file exists but we can't read it,
	//  which might be caused by inappropriate use of sudo
	if t, err := core.ConfigFromFile(pwd); err == nil {
		config = t
	}

	// TODO generalize this for other virt engines (maybe)
	machine := vmware.New(config)

	switch command {
	case "":
		return Help(ProgramHelp)
	case "-h":
		return Help(ProgramHelp)
	case "--help":
		return Help(ProgramHelp)
	case "help":

	case "clone":
		source, err := ParseClone(args, config)
		if err != nil {
			return err
		}
		if err := machine.Clone(source); err != nil {
			return err
		}
	case "start":
		if err := machine.Start(); err != nil {
			return err
		}
		fmt.Printf("machine %q running (%s)\n", config.Path, config.Engine)
	case "stop":
		if err := machine.Stop(); err != nil {
			return err
		}
	case "restart":
		if err := machine.Restart(); err != nil {
			return err
		}
	case "ssh":
		if err := SSH(args, machine); err != nil {
			return err
		}
	case "ip":
		ip, err := machine.IP()
		if err != nil {
			return err
		}
		fmt.Println(ip)
	case "mount":
		if err := ParseMounts(args, config); err != nil {
			return err
		}
		if err := machine.Mount(); err != nil {
			return err
		}
	case "delete":
		if err := machine.Delete(); err != nil {
			return err
		}
	default:
		fmt.Printf("Unknown command %q\n\n", command)
		fmt.Print(CommandList)
		return nil
	}

	// If the command ran successfully we'll save and update the machine file.
	// If there was an error earlier we should have aborted already and we'll
	// leave the machine file alone.
	if err := config.Save(pwd); err != nil {
		return fmt.Errorf("error writing changes to %s: %s", core.MachineFile, err)
	}

	return nil
}
