package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cbednarski/lovm/core"
)

func CommandExists(command string) bool {
	err := exec.Command("vmrun").Run()
	if err != nil && strings.Contains(err.Error(), exec.ErrNotFound.Error()) {
		return false
	}
	return true
}

func Prefix(ok bool) string {
	if ok {
		return " √ "
	}
	return " - "
}

type Check func() (bool, string)

// Status displays information about the current system (and possibly currently-
// running VMs). It may suggest but should not make any changes to the user's
// system.
//
// If we fail to detect something we will not error because that is what the
// status command is supposed to do.
//
// Initially it will:
//
// 1. Check whether VMware is installed
// 2. Check whether VirtualBox is installed
// 3. Check whether VirtualBox host-only network has been created (by default
//    it is not) and give instructions on how to create it if it does not exist
func Status(machine core.VirtualizationEngine) error {
	checklist := []Check{
		func() (bool, string) {
			if CommandExists("vmrun") {
				return true, "VMware installed"
			}
			return false, "VMware not found (vmrun)"
		},
		func() (bool, string) {
			if CommandExists("vboxmanage") {
				return true, "VirtualBox installed"
			}
			return false, "VirtualBox not found (vboxmanage)"
		},
		func() (bool, string) {
			if CommandExists("vboxmanage") {
				out, err := exec.Command("vboxmanage", "list", "hostonlyifs").Output()
				if err == nil && !bytes.Equal(out, []byte{}) {
					return true, "VirtualBox host-only interface detected"
				}
				return false, "VirtualBox host-only interface is missing. SSH will not work. See 'lovm help virtualbox'"
			}
			return false, ""
		},
	}

	for _, item := range checklist {
		ok, message := item()
		if message != "" {
			fmt.Println(Prefix(ok), message)
		}
	}

	return core.ErrNotImplemented
}
