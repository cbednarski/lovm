package engine

import (
	"testing"

	"github.com/cbednarski/lovm/core"
	"github.com/cbednarski/lovm/engine/unknown"
	"github.com/cbednarski/lovm/engine/virtualbox"
	"github.com/cbednarski/lovm/engine/vmware"
)

// TestInterfaces provides a compiler check to verify that each engine satisfies
// the expected interfaces.  This is mainly useful when implementing a new
// interface before it is integrated into the Identify function or exposed via
// the UI.
func TestInterfaces(t *testing.T) {
	var verify = func(engine core.VirtualizationEngine) {}

	dummy := &core.MachineConfig{}

	verify(vmware.New(dummy))
	verify(virtualbox.New(dummy))
	verify(unknown.New(dummy))
}

func TestIdentify(t *testing.T) {
	cases := map[string]string{
		"/path/to/some.vmx":            vmware.Identifier,
		"/path/to/some.vmx:snapshotID": vmware.Identifier,
		"/path/to/some.vbox":           virtualbox.Identifier,
		"/path/to/something.else":      unknown.Identifier,
	}

	for input, expected := range cases {
		output := Identify(input)
		if output != expected {
			t.Errorf("Expected %q, found %q", expected, output)
		}
	}
}

func TestEngine(t *testing.T) {
	dummy := &core.MachineConfig{}

	cases := map[core.VirtualizationEngine]string{
		Engine("/path/to/some.vmx", dummy):       vmware.Identifier,
		Engine("/path/to/some.vbox", dummy):      virtualbox.Identifier,
		Engine("/path/to/something.else", dummy): unknown.Identifier,
	}

	for vm, expected := range cases {
		if vm.Type() != expected {
			t.Errorf("Expected %s, found %s", expected, vm.Type())
		}
	}
}
