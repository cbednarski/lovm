package cli

import (
	"os"
	"os/exec"

	"github.com/cbednarski/lovm/core"
)

// SSHBoolFlags are used to identify which flags may be passed to SSH without
// a second argument. This is a heuristic, and may vary based on the SSH
// implementation so we should allow this to be overridden.
//
// TODO make this configurable and/or add defaults from different platforms
const SSHBoolFlags = `46AaCfGgKkMNnqsTtVvXxYy`
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

// SplitSSHRemoteCommands is used to identify parts of the SSH command that are
// expected to be executed on the remote side, after the connection has been
// established. Because we are inserting the hostname (IP address) we have to
// make sure to insert it before any remote commands.
//
// The heuristics we are using are as follows:
//
// 1. Any boolean flag(s) matching SSHBoolFlags are identified and passed
//    through as-is
// 2. Any other flag (starting with -) is assumed to have an argument with no
//    spaces that follows immediately
// 3. Any other argument that does not start with - is assumed to be a remote
//    command, and any additional arguments that follow are assumed to be part
//    of the remote command
//
// Technically we can be more specific than #2 by specifying the full list of
// SSH options, but there are a lot of them and they may vary between systems,
// so we'll use the heuristics instead.
//
// The first []string returned contains all SSH flags and options, if they
// exist. The second []string returned contains all remote command components,
// if they exist.
func SplitSSHRemoteCommands(args []string) (sshArgs []string, remoteCommands []string) {
	splitIndex := -1

	for i := 0; i < len(args); i++ {
		if IsSSHBoolFlags(args[i]) {
			sshArgs = append(sshArgs, args[i])
		} else if IsSSHOptionFlag(args[i]) && i+1 < len(args) {
			sshArgs = append(sshArgs, args[i], args[i+1])
		} else {
			splitIndex = i
			break
		}
	}

	if splitIndex < len(args) {
		remoteCommands = append(args[splitIndex:])
	}

	return
}

func SSH(args []string, machine core.VirtualizationEngine) error {
	ip, err := machine.IP()
	if err != nil {
		return err
	}

	// Any additional arguments (-i, -l, etc.) may be passed through to the
	// underlying ssh command, while the IP is filled in automatically
	args = append(args, ip.String())

	command := exec.Command("ssh", args...)

	// Pass through stdin, stdout, and stderr to the child process
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

