package core

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

const MachineFile = "machine.lovm"

// MachineConfig represents a cloned VM and the information we need to find it
// or re-clone it from scratch after a delete operation is called on the clone.
type MachineConfig struct {
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
