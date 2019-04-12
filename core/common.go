package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrNotImplemented = errors.New("not implemented")

func CommandError(command *exec.Cmd, output []byte) {
	os.Stderr.WriteString(fmt.Sprintf("[command debug] %s\n", strings.Join(command.Args, " ")))
	os.Stderr.Write(output)
}
