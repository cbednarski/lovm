package virtualbox

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cbednarski/lovm/core"
)

const SnapshotName = `lovm-clone`

type VirtualBox struct {
	Config *core.MachineConfig
}

func New(config *core.MachineConfig) *VirtualBox {
	return &VirtualBox{
		Config: config,
	}
}

func (v *VirtualBox) Clone(source string) error {
	if source == "" && v.Config.Source == "" {
		return errors.New("clone a virtual machine first")
	}

	if v.Found() {
		// If the VM is already cloned but we've been asked to clone a
		// different source than the one we cloned, error and inform the user
		// that they need to destroy first
		if source != "" && source != v.Config.Source {
			return fmt.Errorf("asked to clone from %q but the virtual "+
				"machine is already cloned from %q; run delete first", source,
				v.Config.Source)
		}
		// If the VM is already cloned and the source is the same it's a no-op
		return nil
	}

	// If there is no user input use the same source they entered earlier
	if source == "" {
		source = v.Config.Source
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

	// We'll create the VM name based on the pwd. Not perfect, but good enough.
	_, targetName := filepath.Split(pwd)

	// VirtualBox automatically implies the name of the containing folder from
	// the name of the VM rather than taking the folder name as input, so we
	// need to keep track of all three parts here.
	//
	// basefolder    .lovm
	// name          example
	// target        .lovm/example/example.vbox
	//
	// This only matters for cloning. Once we have the .vbox file in place and
	// it has been registered we can simply reference it by path.
	baseFolder := filepath.Join(pwd, ".lovm")
	target := filepath.Join(baseFolder, targetName,
		fmt.Sprintf("%s.vbox", targetName))

	args := []string{"clonevm", source, "--options", "link",
		"--basefolder", baseFolder, "--name", targetName, "--register"}

	if snapshot == "" {
		// A snapshot is required for a linked clone in VirtualBox, so we'll
		// create one if the user didn't specify anything.
		if err := CreateSnapshot(source); err != nil {
			return errors.New("failed to create snapshot required for cloning")
		}

		// We're going to use the special snapshot but we do not write it into
		// the config struct because the user did not explicitly specify it.
		args = append(args, `--snapshot`, SnapshotName)
	} else {
		args = append(args, `--snapshot`, snapshot)
	}

	cmd := exec.Command("vboxmanage", args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		core.CommandError(cmd, out)
	}

	// Set VM path to the .vbox file we just created.
	v.Config.Path = target
	v.Config.Source = source
	if snapshot != "" {
		v.Config.Source = fmt.Sprintf("%s:%s", source, snapshot)
	}

	return err
}

func (v *VirtualBox) Start() error {
	if err := v.Clone(""); err != nil {
		return err
	}

	// TODO VirtualBox barfs if we try to start a VM that is already running so
	//  we'll need to detect that case and no-op instead.

	cmd := exec.Command("vboxmanage",
		"startvm", v.Config.Path, "--type", "headless")

	out, err := cmd.CombinedOutput()

	if err != nil {
		core.CommandError(cmd, out)
	}

	return err
}

func (v *VirtualBox) Stop() error {
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	cmd := exec.Command("vboxmanage",
		"controlvm", v.Config.Path, "poweroff")

	out, err := cmd.CombinedOutput()

	if err != nil {
		// If the error message says the VM is already turned off, we're done
		if bytes.Contains(out, []byte(`is not currently running`)) {
			return nil
		}

		core.CommandError(cmd, out)
	}

	return err
}

func (v *VirtualBox) Restart() error {
	if err := v.Stop(); err != nil {
		return err
	}
	if err := v.Start(); err != nil {
		return err
	}
	return nil
}

func (v *VirtualBox) Delete() error {
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	if err := v.Stop(); err != nil {
		return err
	}

	cmd := exec.Command("vboxmanage",
		"unregistervm", v.Config.Path, "--delete")

	out, err := cmd.CombinedOutput()

	if err != nil {
		core.CommandError(cmd, out)
	}

	// Remove the machine path because we don't have a VM anymore
	if err == nil {
		v.Config.Path = ""
	}

	return err
}

func (v *VirtualBox) IP() (net.IP, error) {
	return nil, core.ErrNotImplemented
}

func (v *VirtualBox) Mount() error {
	return core.ErrNotImplemented
}

func (v *VirtualBox) Found() bool {
	if v.Config.Path == "" {
		return false
	}

	fi, err := os.Stat(v.Config.Path)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

func DetectSnapshot(path string) bool {
	cmd := exec.Command("vboxmanage",
		"snapshot", path, "showvminfo", SnapshotName)

	err := cmd.Run()

	// If we get any kind of error that means the snapshot doesn't exist
	return err == nil
}

func CreateSnapshot(path string) error {
	// If the snapshot already exists don't make another one, because that would
	// be silly
	if DetectSnapshot(path) {
		return nil
	}

	cmd := exec.Command("vboxmanage",
		"snapshot", path, "take", SnapshotName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		core.CommandError(cmd, out)
		return err
	}

	// This will only happen the first time we clone a particular VM, so we'll
	// let the user know what's happening.
	fmt.Printf("created snapshot %q for %q", SnapshotName, path)

	return nil
}
