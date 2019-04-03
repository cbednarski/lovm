package engine

import (
	"strings"

	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine/vmware"
)

const (
	Unknown = "unknown"
	VMware  = "vmware"
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
	return Unknown
}

// Engine returns an implementation of lovm.VirtualizationEngine based on the
// type of engine determined by Identify
func Engine(machine *core.MachineConfig) core.VirtualizationEngine {
	switch Identify(machine.Source) {
	case VMware:
		return vmware.New(machine)
	}
	return nil
}
