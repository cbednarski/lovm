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

func TestReadMACAdressesFromVMX_OK(t *testing.T) {
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

func TestReadMACAdressesFromVMX_NoNIC(t *testing.T) {
	_, err := ReadMACAdressesFromVMX(filepath.Join("test-fixtures", "centos-nonic.vmx"))
	if err != ErrInterfaceNotFound {
		t.Errorf("Expected %s, found %s", ErrInterfaceNotFound, err)
	}
}

func TestFindCurrentLeaseByMAC_OK(t *testing.T) {
	path := filepath.Join("test-fixtures", "dhcpd%d.leases")

	mac, err := net.ParseMAC("00:0c:29:f7:07:f2")
	if err != nil {
		t.Fatal(err)
	}

	expected := "172.16.23.131"
	ip, err := FindCurrentLeaseByMAC(path, 8, mac)
	if err != nil {
		t.Fatal(err)
	}
	if ip.String() != expected {
		t.Errorf("Expected %s, found %s", expected, ip.String())
	}
}

func TestFindCurrentLeaseByMAC_MACNotFound(t *testing.T) {
	path := filepath.Join("test-fixtures", "dhcpd%d.leases")

	// This MAC address does not exist in the lease file at all
	mac, err := net.ParseMAC("aa:0c:29:ce:c0:a8")
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindCurrentLeaseByMAC(path, 8, mac)
	if err != ErrLeaseNotFound {
		t.Errorf("Expected %s, got %s", ErrLeaseNotFound, err)
	}
}

func TestFindCurrentLeaseByMAC_NoCurrentLease(t *testing.T) {
	path := filepath.Join("test-fixtures", "dhcpd%d.leases")

	// This MAC address has a lease but it has expired
	mac, err := net.ParseMAC("00:0c:29:ce:c0:a8")
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindCurrentLeaseByMAC(path, 8, mac)
	if err != ErrLeaseNotFound {
		t.Errorf("Expected %s, got %s", ErrLeaseNotFound, err)
	}
}

func TestDetectIPFromMACAddress(t *testing.T) {
	networkConfigFile := filepath.Join("test-fixtures", "networking")
	dhcpLeasesFile := filepath.Join("test-fixtures", "dhcpd%d.leases")

	mac, err := net.ParseMAC("00:0c:29:f7:07:f2")
	if err != nil {
		t.Fatal(err)
	}

	ip, err := DetectIPFromMACAddress(networkConfigFile, dhcpLeasesFile, mac)
	if err != nil {
		t.Fatal(err)
	}

	expected := "172.16.23.131"
	if ip.String() != expected {
		t.Errorf("Expected %s, found %s", expected, ip.String())
	}
}
