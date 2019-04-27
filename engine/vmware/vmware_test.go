package vmware

import (
	"net"
	"path/filepath"
	"testing"
)

func CompareIntLists(a, b []int) bool {
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

func TestListDHCPVirtualNetworks(t *testing.T) {
	networks, err := ListDHCPVirtualNetworks(filepath.Join("test-fixtures", "networking"))
	if err != nil {
		t.Fatal(err)
	}

	expected := []int{1, 8}

	if !CompareIntLists(networks, expected) {
		t.Errorf("Expected %#v, found %#v", expected, networks)
	}
}

func TestReadMACAdressesFromVMX(t *testing.T) {
	macs, err := ReadMACAdressesFromVMX(filepath.Join("test-fixtures", "centos.vmx"))
	if err != nil {
		t.Fatal(err)
	}

	expected, err := net.ParseMAC("00:0c:29:f7:07:f2")
	if err != nil {
		t.Fatal(err)
	}

	actual, ok := macs["ethernet0"]
	if !ok {
		t.Fatal("Could not find interface named ethernet0")
	}

	if actual.String() != expected.String() {
		t.Errorf("Expected %q, found %q", expected.String(), actual.String())
	}
}

func TestFindCurrentLeaseByMAC(t *testing.T) {
	path := filepath.Join("test-fixtures", "dhcpd.leases")

	mac, err := net.ParseMAC("00:0c:29:f7:07:f2")
	if err != nil {
		t.Fatal(err)
	}

	ip, err := FindCurrentLeaseByMAC(path, 8, mac)
}
