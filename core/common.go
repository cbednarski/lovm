package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	// ErrNotImplemented is returned as a placeholder when a particular feature
	// is not yet implemented. This allows any part of the implementation to
	// satisfy interface requirements before a full implementation is available.
	ErrNotImplemented = errors.New("not implemented (TODO)")
)

func CommandError(command *exec.Cmd, output []byte) {
	os.Stderr.WriteString(fmt.Sprintf("[command debug] %s\n", strings.Join(command.Args, " ")))
	os.Stderr.Write(output)
}
