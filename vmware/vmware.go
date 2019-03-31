package vmware

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cbednarski/lovm/lovm"
)

type VMware struct {
	VM *lovm.VirtualMachine
}

func New(vm *lovm.VirtualMachine) *VMware {
	return &VMware{
		VM: vm,
	}
}

func (v *VMware) Clone(source string) error {
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

func (v *VMware) Start() error {
	cmd := exec.Command("vmrun", "start", v.VM.Path, "nogui")

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	return err
}
