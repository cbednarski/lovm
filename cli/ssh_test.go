package cli

import (
	"net"
	"strings"
	"testing"
)

func TestIsSSHBoolFlags(t *testing.T) {
	cases := map[string]bool{
		"a":      false,
		"-a":     true,
		"-Aa":    true,
		"-b":     false,
		"-c":     false,
		"-MN6XY": true,
		"MN6XY":  false,
	}

	for input, expected := range cases {
		if IsSSHBoolFlags(input) != expected {
			if expected {
				t.Errorf("Expected %s to be a bool flag", input)
			} else {
				t.Errorf("Expected %s NOT to be a bool flag", input)
			}
		}
	}
}

func TestIsSSHOptionFlag(t *testing.T) {
	cases := map[string]bool{
		"-b":  true,
		"-c":  true,
		"c":   false,
		"-bc": false,
		"-q":  false,
	}

	for input, expected := range cases {
		if IsSSHOptionFlag(input) != expected {
			if expected {
				t.Errorf("Expected %s to be an option flag", input)
			} else {
				t.Errorf("Expected %s NOT to be an option flag", input)
			}
		}
	}
}

func CompareLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	for index, word := range a {
		if word != b[index] {
			return false
		}
	}
	return true
}

func TestSplitSSHRemoteCommands(t *testing.T) {
	cases := map[string][][]string{
		``: { // Empty string
			{},
			{},
		},
		`sudo shutdown -r now`: {
			{},
			{"sudo", "shutdown", "-r", "now"},
		},
		`-i ~/.ssh/id_rsa`: {
			{"-i", "~/.ssh/id_rsa"},
			{},
		},
		`-l root shutdown -r now`: {
			{"-l", "root"},
			{"shutdown", "-r", "now"},
		},
	}

	for input, expected := range cases {
		inputs := strings.Split(input, " ")
		if input == "" {
			inputs = []string{}
		}
		sshArgs, remoteCommands := SplitSSHRemoteCommands(inputs)
		if !CompareLists(sshArgs, expected[0]) {
			t.Errorf("input %q had unexpected sshArgs: %#v (expected %#v)", input, sshArgs, expected[0])
		}
		if !CompareLists(remoteCommands, expected[1]) {
			t.Errorf("input %q had unexpected remoteCommands: %#v (expected %#v)", input, remoteCommands, expected[1])
		}
	}
}

func TestBuildSSHCommand(t *testing.T) {
	type testCase struct {
		Args     []string
		IP       net.IP
		Expected []string
	}

	cases := []testCase{
		{
			Args:     []string{"-l", "root"},
			IP:       net.ParseIP("192.168.1.80"),
			Expected: []string{"ssh", "-l", "root", "192.168.1.80"},
		},
		{
			Args:     []string{"shutdown", "-h", "now"},
			IP:       net.ParseIP("192.168.1.80"),
			Expected: []string{"ssh", "192.168.1.80", "shutdown", "-h", "now"},
		},
		{
			Args:     []string{"-l", "root", "shutdown", "-h", "now"},
			IP:       net.ParseIP("192.168.1.80"),
			Expected: []string{"ssh", "-l", "root", "192.168.1.80", "shutdown", "-h", "now"},
		},
	}

	for _, c := range cases {
		output := BuildSSHCommand(c.Args, c.IP)
		if !CompareLists(c.Expected, output.Args) {
			t.Errorf("Expected command %v, found %v", c.Expected, output.Args)
		}
	}
}
