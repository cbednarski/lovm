package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.stormbase.io/cbednarski/cli"
	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine"
)

var topics = map[string]string{}

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

const Footer = `
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

	config, err := ConfigFromFileOrNew(pwd)
	if err != nil {
		return err
	}

	machine := engine.Engine(config.Source, config)

	commands := map[string]*cli.Command{
		"clone": {
			Summary: "Clone a VM. Start here!",
			Run: func(args []string) error {
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
				return nil
			},
		},
		"start": {
			Summary: "Start the VM",
			Run: func(args []string) error {
				if err := machine.Start(); err != nil {
					return err
				}
				fmt.Printf("machine %q running (%s)\n", config.Path, config.Engine)
				return nil
			},
		},
		"stop": {
			Run: func(args []string) error {
				return machine.Stop()
			},
		},
		"restart": {
			Run: func(args []string) error {
				return machine.Restart()
			},
		},
		"ssh": {
			Run: func(args []string) error {
				return SSH(args, machine, config)
			},
		},
		"ip": {
			Run: func(args []string) error {
				ip, err := machine.IP()
				if err != nil {
					return err
				}
				fmt.Println(ip)
				return nil
			},
		},
		"mount": {
			Run: func(args []string) error {
				if err := ParseMounts(args, config); err != nil {
					return err
				}
				if err := machine.Mount(); err != nil {
					return err
				}
				return nil
			},
		},
		"delete": {
			Run: func(args []string) error {
				return machine.Delete()
			},
		},

	}

	app := &cli.CLI{
		Name: "lovm",
		Header: "A minimalist, idempotent command-line tool for managing local virtual machines",
		Version: "0.1.0",
		Footer: Footer,
		Commands: commands,
		HelpTopics: topics,
	}

	if err := app.Run(); err != nil {
		return err
	}

	// If the command ran successfully we'll save and update the machine file.
	// If there was an error earlier we should have aborted already and we'll
	// leave the machine file alone.
	if err := config.Save(pwd); err != nil {
		return fmt.Errorf("error writing changes to %s: %s", core.MachineFile, err)
	}

	return nil
}
