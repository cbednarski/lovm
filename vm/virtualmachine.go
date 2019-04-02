package vm

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"path/filepath"
)

const MachineFile = "machine.lovm"

// VirtualMachine represents a cloned VM and the information we need to find it
// or re-clone it from scratch after a delete operation is called on the clone.
type VirtualMachine struct {
	// Path to the current VM; may be a file (.vmx) or something else, depending
	// on the implementation of the virtualization engine.
	Path string `json:"path"`

	// Mounts contains a map of paths on the host filesystem to be mounted into
	// the guest filesystem
	Mounts map[string]string `json:"mounts"`

	// Source indicates the location of the parent VM (i.e. the VM we cloned to
	// make this one). A colon (:) is used to specify a snapshot.
	Source string `json:"source"`

	// Engine is used to cache the virtualization engine used for this VM
	Engine string `json:"engine"`
}

// VirtualizationEngine abstracts various ways a virtualization engine might do
// things.
//
// When implementing an engine, remember: The user told the machine to do
// something NOW so don't wait for the guest OS to cooperate.
type VirtualizationEngine interface {
	// Clone clones the VM. If the VM is already cloned, do nothing. If the VM
	// is already cloned but the user has tried to change the source complain.
	Clone(string) error

	// Start the VM. If the VM is not cloned but we know how (machine.lovm is
	// already populated), clone it and start it. If the machine is in a weird
	// state (paused or something, start it). If the machine is already running
	// just report a success. If the machine can't be started, tell the user
	// why not.
	Start() error

	// Stop the VM. If it won't do a clean shutdown, just cut the power or kill
	// it. Fast. The user said stop to the machine, not to the OS, so stop the
	// machine. If they want to do a clean shutdown they can do it inside the
	// OS. You write crash-only software, don't you? Stop should not take more
	// than 10 seconds.
	Stop() error

	// Restart the VM. Stop. Then start. If they want to do shutdown -r now they
	// can do it via SSH. If the VM is not actually running just start it up.
	Restart() error

	// Delete the VM. If it's running, stop it first and then delete it. Don't
	// delete the machine.lovm file. Just the VM files. This command is the Nuke
	// It From Orbit button and should always succeed in the most expeditious
	// way possible. The user said Delete so they don't care if things are saved
	// before shutdown. kill -9 and rm -rf if you have to.
	Delete() error

	// IP is used for SSH, or sometimes just to show the user the IP. Pick the
	// first one because that's probably what they want. Maybe we'll get fancy
	// later and support multiple IPs, but not now.
	IP() (net.IP, error)

	// Mount shared folders. When the user runs the mount command they will add
	// a mount to the list of mounts, so this implementation needs to figure
	// out three things: Enable mounting things. Did we already mount that
	// thing? Mount the thing.
	//
	// In VMware, for example, mounts need to be re-enabled each time the
	// machine is started or rebooted. The user doesn't need to know this. Just
	// mount the thing.
	Mount() error

	// Found returns true if the VM already exists
	Found() bool
}

// VirtualMachineFromFile looks for a file called "machine.lovm" in the
// specified path, and returns a VirtualMachine if it finds one.
func VirtualMachineFromFile(path string) (*VirtualMachine, error) {
	filename := filepath.Join(path, MachineFile)

	vm := &VirtualMachine{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, vm); err != nil {
		return nil, err
	}

	return vm, nil
}

// Save writes the VM's configuration to a file called "machine.lovm" in the
// specified folder.
func (v *VirtualMachine) Save(path string) error {
	filename := filepath.Join(path, MachineFile)

	data, err := json.MarshalIndent(v, "", "  ")
	// add a newline to the end of the file so we can inspect it with cat
	// without screwing up the terminal
	data = append(data, byte(10))
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	return nil
}
