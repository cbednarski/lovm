package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine"
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
	// configured in machine.lovm (if it exists). If machine.Source is empty and
	// there is no user input, we'll complain.
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

func ConfigFromFileOrNew(path string) (*core.MachineConfig, error) {
	config, err := core.ConfigFromFile(path)
	if err != nil {
		// If the error specifically says that the file does not exist then we
		// will simply create a new, empty config and move on, because that's
		// what we would do anyway. If there is a different type of io error,
		// such as we can't read the file, or there is a problem parsing the
		// JSON, then we'll show that error to the user
		if strings.Contains(err.Error(), "no such file or directory") {
			return &core.MachineConfig{}, nil
		}
		return nil, err
	}
	return config, nil
}

// Text allows you to return a nil error after showing help text
func Text(text string) error {
	fmt.Print(text)
	return nil
}

func Main() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	command, args := ParseArgs(os.Args)

	config, err := ConfigFromFileOrNew(pwd)
	if err != nil {
		return err
	}

	machine := engine.Engine(config.Source, config)

	switch command {
	case "":
		return Text(ProgramHelp)
	case "-h":
		return Text(ProgramHelp)
	case "--help":
		return Text(ProgramHelp)
	case "help":
		// TODO add interactive help command here for different commands
		return Help(args)
	case "clone":
		source, err := ParseClone(args, config)
		if err != nil {
			return err
		}
		// Override the engine type if there is CLI input, because the user
		// might be cloning a different type of VM after deleting a previous one
		if source != "" {
			machine = engine.Engine(source, config)
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
		fmt.Print(CommandList)
		return fmt.Errorf("unrecognized command %q", command)
	}

	// If the command ran successfully we'll save and update the machine file.
	// If there was an error earlier we should have aborted already and we'll
	// leave the machine file alone.
	if err := config.Save(pwd); err != nil {
		return fmt.Errorf("error writing changes to %s: %s", core.MachineFile, err)
	}

	return nil
}
