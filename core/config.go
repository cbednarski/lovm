package core

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

const MachineFile = "machine.lovm"

// SSHConfig is used to store additional details for SSHing into the virtual
// machine, including a way to match the network interface name / ID, SSH
// username, ssh private key, and possibly other configuration
type SSHConfig struct {
	// Login may be used to specify the unix login to use for SSH (e.g. "root").
	// This will fill in the -L option for the underlying SSH command
	Login string `json:"login,omitempty"`

	// PrivateKeyPath may be used to specify a private key to use for SSH. It
	// will fill in the -i option for the underlying SSH command
	PrivateKeyPath string `json:"private-key-path,omitempty"`

	// NetworkInterface may be specified if the virtual machine has more than
	// one NIC. The virtual machine engine can use this to match the MAC
	// address. See the specific engine for details.
	NetworkInterface string `json:"network-interface,omitempty"`
}

// MachineConfig represents a cloned VM and the information we need to find it
// or re-clone it from scratch after a delete operation is called on the clone.
type MachineConfig struct {
	// Path to the current VM; may be a file (.vmx) or something else, depending
	// on the implementation of the virtualization engine.
	Path string `json:"path,omitempty"`

	// Mounts contains a map of paths on the host filesystem to be mounted into
	// the guest filesystem
	Mounts map[string]string `json:"mounts,omitempty"`

	// Source indicates the location of the parent VM (i.e. the VM we cloned to
	// make this one). A colon (:) is used to specify a snapshot.
	Source string `json:"source"`

	// Engine is used to cache the virtualization engine used for this VM
	Engine string `json:"engine,omitempty"`

	// SSH stores some additional configuration for the ssh command
	SSH SSHConfig `json:"ssh-config,omitempty"`
}

// ConfigFromFile looks for a file called "machine.lovm" in the specified path,
// and returns a MachineConfig if it finds one.
func ConfigFromFile(path string) (*MachineConfig, error) {
	filename := filepath.Join(path, MachineFile)

	vm := &MachineConfig{}

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
func (c *MachineConfig) Save(path string) error {
	filename := filepath.Join(path, MachineFile)

	data, err := json.MarshalIndent(c, "", "  ")
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
