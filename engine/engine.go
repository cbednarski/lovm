package engine

import (
	"strings"

	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine/unknown"
	"github.com/cbednarski/lovm/engine/virtualbox"
	"github.com/cbednarski/lovm/engine/vmware"
)

const (
	Unknown    = "unknown"
	VMware     = "vmware"
	VirtualBox = "virtualbox"
)

// Identify uses heuristics to determine the appropriate virtualization engine
// for the specified source
func Identify(source string) string {
	// We use the : separator as a special case for indicating snapshots. This
	// breaks the filename heuristic so we'll strip it off first.
	if strings.Contains(source, ":") {
		source = strings.Split(source, ":")[0]
	}

	if strings.HasSuffix(source, ".vmx") {
		return VMware
	}
	if strings.HasSuffix(source, ".vbox") {
		return VirtualBox
	}
	return Unknown
}

// Engine returns an implementation of lovm.VirtualizationEngine based on the
// type of engine determined by Identify
func Engine(source string, machine *core.MachineConfig) core.VirtualizationEngine {
	switch Identify(source) {
	case VMware:
		return vmware.New(machine)
	case VirtualBox:
		return virtualbox.New(machine)
	}
	return unknown.New(machine)
}
