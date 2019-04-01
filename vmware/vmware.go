package vmware

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cbednarski/lovm/vm"
)

var errNotImplemented = errors.New("not implemented")

type VMware struct {
	VM *vm.VirtualMachine
}

func New(vm *vm.VirtualMachine) *VMware {
	return &VMware{
		VM: vm,
	}
}

func (v *VMware) Clone(source string) error {
	if source == "" && v.VM.Source == "" {
		return errors.New("clone a VM first")
	}

	// TODO Possible scenarios:
	//  1. machine.lovm does not exist. User has not specified anything. Error.
	//  2. machine.lovm does not exist. User has specified something. Continue.
	//  3. machine.lovm exists. Path does not exist. User specified nothing. Continue.
	//  4. machine.lovm exists. Path does not exist. User specified something. Continue.
	//  5. machine.lovm exists. Path exists. User specified same source. Continue (no-op)
	//  6. machine.lovm exists. Path exists. User specified different source. Error.

	// Make a subfolder for the VM files to live in
	if err := os.MkdirAll(".lovm", 0755); err != nil {
		return err
	}

	snapshot := ""
	if strings.Contains(source, ":") {
		snapshot = strings.Split(source, ":")[1]
		source = strings.Split(source, ":")[0]
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// We'll create the VMX name based on the pwd. Not perfect, but good enough.
	_, targetName := filepath.Split(pwd)

	target := filepath.Join(pwd, ".lovm", targetName, fmt.Sprintf("%s.vmx", targetName))

	args := []string{"clone", source, target, "linked"}

	if snapshot != "" {
		args = append(args, fmt.Sprintf("-snapshot=%s", snapshot))
	}

	cmd := exec.Command("vmrun", args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	// Set VM path to the vmx file we just created.
	v.VM.Path = target

	return err
}

// Found will check for the presence of a vmx file
func (v *VMware) Found() bool {
	fi, err := os.Stat(v.VM.Path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func (v *VMware) Start() error {
	if err := v.Clone(""); err != nil {
		return err
	}

	cmd := exec.Command("vmrun", "start", v.VM.Path, "nogui")

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	return err
}

func (v *VMware) Stop() error {
	cmd := exec.Command("vmrun", "stop", v.VM.Path, "hard")

	out, err := cmd.CombinedOutput()

	if err != nil {
		// If the error message says the VM is already turned off, just say
		// we're done, no error.
		if bytes.Contains(out, []byte(`The virtual machine is not powered on`)) {
			return nil
		}

		log.Printf("%s", out)
	}

	return err
}

func (v *VMware) Restart() error {
	if err := v.Stop(); err != nil {
		return err
	}
	if err := v.Start(); err != nil {
		return err
	}
	return nil
}

//TODO implement IP
// IP returns the first IP address associated with the virtual machine. There
// may be more than one. This is currently unhandled / undefined behavior.
//
// See https://www.vmware.com/support/ws55/doc/ws_net_advanced_ipaddress.html
// for how this is implemented on Linux.
func (v *VMware) IP() (*net.IP, error) {
	return nil, errNotImplemented
}

func (v *VMware) Delete() error {
	if err := v.Stop(); err != nil {
		return err
	}

	cmd := exec.Command("vmrun", "deleteVM", v.VM.Path)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	return err
}

// TODO implement Mount
func (v *VMware) Mount() error {
	// TODO check guest tools status because shared folders don't work without
	//  those
	return errNotImplemented
}
