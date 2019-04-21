package cli

import (
	"net"
	"os"
	"os/exec"

	"github.com/cbednarski/lovm/core"
)

// SSHBoolFlags are used to identify which flags may be passed to SSH without
// a second argument. This is a heuristic.
const SSHBoolFlags = `46AaCfGgKkMNnqsTtVvXxYy`

// SSHOptionFlags are used to identify options that may be passed to SSH with
// a single additional argument (no spaces). This is a heuristic.
const SSHOptionFlags = `bcDEeFIiJLlmooOopQrRsSttWw`

// IsSSHBoolFlags checks whether the specified flag contains one or more SSH
// boolean flags. For example, something in the form of ssh -4 or ssh -A6
// We do not validate these, we merely determine whether they are boolean flags
// or whether they are options.
func IsSSHBoolFlags(arg string) bool {
	if len(arg) < 2 {
		return false
	}
	if arg[0] != '-' {
		return false
	}

	for _, a := range []byte(arg[1:]) {
		found := false
		for _, s := range []byte(SSHBoolFlags) {
			if a == s {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func IsSSHOptionFlag(arg string) bool {
	if len(arg) != 2 {
		return false
	}
	if arg[0] != '-' {
		return false
	}

	for _, s := range []byte(SSHOptionFlags) {
		if arg[1] == s {
			return true
		}
	}
	return false
}

// SplitSSHRemoteCommands identifies options and flags passed to SSH from any
// command to be executed on the remote side. lovm ssh works by inserting the
// virtual machine's IP address but SSH requires that all flags and options be
// passed before the hostname argument, and any subsequent commands are passed
// to the remote shell to be executed, so we need to be able to insert the IP
// address *between* the SSH args and remote commands.
//
// For example:    lovm ssh -l root shutdown -h now
// Translates to:  ssh -l root <ip> shutdown -h now
//
// The heuristics we are using are as follows:
//
// 1. Any boolean flag(s) matching SSHBoolFlags are identified and passed along
// 2. Any option flag (starting with -) is assumed to have an argument with no
//    spaces that follows immediately
// 3. Any argument that does not start with - and does not satisfy #2 above
//    marks the start of the remote command along with any subsequent arguments
//
// The first []string returned contains all SSH flags and options, if any are
// present. The second []string returned contains all remote command components,
// if any are present.
func SplitSSHRemoteCommands(args []string) (sshArgs []string, remoteCommands []string) {
	splitIndex := -1

	if len(args) == 0 {
		return
	}

	for i := 0; i < len(args); i++ {
		if IsSSHBoolFlags(args[i]) {
			sshArgs = append(sshArgs, args[i])
		} else if IsSSHOptionFlag(args[i]) && i+1 < len(args) {
			sshArgs = append(sshArgs, args[i], args[i+1])
			i++
		} else {
			splitIndex = i
			break
		}
	}

	if -1 < splitIndex && splitIndex < len(args) {
		remoteCommands = append(args[splitIndex:])
	}

	return
}

func BuildSSHCommand(args []string, ip net.IP) *exec.Cmd {
	sshArgs, remoteCommands := SplitSSHRemoteCommands(args)
	finalArgs := append(sshArgs, ip.String())
	finalArgs = append(finalArgs, remoteCommands...)

	return exec.Command("ssh", finalArgs...)
}

func SSH(args []string, machine core.VirtualizationEngine) error {
	ip, err := machine.IP()
	if err != nil {
		return err
	}

	command := BuildSSHCommand(args, ip)

	// Attach stdin, stdout, and stderr to the child process
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Fork the child process
	if err := command.Start(); err != nil {
		return err
	}

	// Wait for it to complete
	if err := command.Wait(); err != nil {
		return err
	}

	return nil
}
