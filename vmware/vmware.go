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
	// Check if we have enough user input to clone something
	if source == "" && v.VM.Source == "" {
		return errors.New("clone a VM first")
	}

	if v.Found() {
		// If the VM is already cloned but we've been asked to clone a
		// different source than the one we cloned, error and inform the user
		// that they need to destroy first
		if source != "" && source != v.VM.Source {
			return fmt.Errorf("asked to clone from %q but vm is already cloned from %q; run destroy first", source, v.VM.Source)
		}
		// If the VM is already cloned and the source is the same it's a no-op
		return nil
	}

	// If there is no user input use the same source they entered earlier
	if source == "" {
		source = v.VM.Source
	}

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
	if v.VM.Path == "" {
		return false
	}

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
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	cmd := exec.Command("vmrun", "stop", v.VM.Path, "hard")

	out, err := cmd.CombinedOutput()

	if err != nil {
		// If the error message says the VM is already turned off, we're done
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
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	if err := v.Stop(); err != nil {
		return err
	}

	cmd := exec.Command("vmrun", "deleteVM", v.VM.Path)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	// Remove the machine path because we don't have a VM anymore
	if err == nil {
		v.VM.Path = ""
	}

	return err
}

// TODO implement Mount
func (v *VMware) Mount() error {
	// TODO check guest tools status because shared folders don't work without
	//  those
	return errNotImplemented
}
