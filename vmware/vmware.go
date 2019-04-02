package vmware

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cbednarski/lovm/vm"
)

const (
	NetworkConfigFile = "/etc/vmware/networking"
	DHCPLeasesFile    = "/etc/vmware/vmnet%d/dhcpd/dhcpd.leases"
	DHCPDateFormat = `2006/01/02 15:04:05`
)

var (
	ErrNotFound        = errors.New("not found")
	errNotImplemented  = errors.New("not implemented")
	reGeneratedAddress = regexp.MustCompile(`(ethernet\d+)\.generatedAddress ?= ?"([0-9a-fA-F:]+)"`)
	reNetworkingConfig = regexp.MustCompile(`answer VNET_(\d+)_DHCP yes`)
	reDHCPLeases       = regexp.MustCompile(`lease ([0-9\.]+) {\s+`+
	                                        `starts [0-9]+ ([0-9/: ]+);\s+`+
	                                        `ends [0-9]+ ([0-9/: ]+);\s+`+
	                                        `hardware ethernet ([0-9a-f:]+);\s+`)
)


type VMware struct {
	VM *vm.VirtualMachine
}

func New(vm *vm.VirtualMachine) *VMware {
	return &VMware{
		VM: vm,
	}
}

func (v *VMware) Clone(source string) error {
	// Check if we have enough user input to clone something
	if source == "" && v.VM.Source == "" {
		return errors.New("clone a VM first")
	}

	if v.Found() {
		// If the VM is already cloned but we've been asked to clone a
		// different source than the one we cloned, error and inform the user
		// that they need to destroy first
		if source != "" && source != v.VM.Source {
			return fmt.Errorf("asked to clone from %q but vm is already cloned from %q; run destroy first", source, v.VM.Source)
		}
		// If the VM is already cloned and the source is the same it's a no-op
		return nil
	}

	// If there is no user input use the same source they entered earlier
	if source == "" {
		source = v.VM.Source
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

// Found will check for the presence of a vmx file
func (v *VMware) Found() bool {
	if v.VM.Path == "" {
		return false
	}

	fi, err := os.Stat(v.VM.Path)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

func (v *VMware) Start() error {
	if err := v.Clone(""); err != nil {
		return err
	}

	cmd := exec.Command("vmrun", "start", v.VM.Path, "nogui")

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	return err
}

func (v *VMware) Stop() error {
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	cmd := exec.Command("vmrun", "stop", v.VM.Path, "hard")

	out, err := cmd.CombinedOutput()

	if err != nil {
		// If the error message says the VM is already turned off, we're done
		if bytes.Contains(out, []byte(`The virtual machine is not powered on`)) {
			return nil
		}

		log.Printf("%s", out)
	}

	return err
}

func (v *VMware) Restart() error {
	if err := v.Stop(); err != nil {
		return err
	}
	if err := v.Start(); err != nil {
		return err
	}
	return nil
}

// IP returns the first IP address associated with the virtual machine. There
// may be more than one. This is currently unhandled / undefined behavior.
func (v *VMware) IP() (net.IP, error) {
	// The following networking concepts will be useful for understanding the
	// implementation which detects the VM's IP address(es). VMware Workstation
	// can create bridged, NATed, or host-only networks, and manages DHCP. Also
	// note that any given VM may have zero or more virtual network devices, and
	// may be part of multiple networks.
	//
	// Terminology
	//
	// - Bridged network masquerades a virtual interface on your physical
	//   networking card. This means that the VM's IP address will appear as a
	//   separate device on your local network. Usually this causes problems on
	//   authenticated networks like wifi so typically VMs use NAT instead. When
	//   using bridged networking, other devices on the network can connect
	//   directly to your VM.
	//
	// - NAT stands for "network address translation" and is commonly used on
	//   routers to route traffic back to devices behind the router. In VMware
	//   all traffic enters via your host's IP and the VMware NAT device sends
	//   it back to the VM. The NAT keeps track of outbound packets and returns
	//   network to the appropriate VM, so your VMs do not appear as separate
	//   devices on the local network.
	//
	// - Host-only networks exist only on your computer, and VMs that are
	//   connected to them cannot reach the internet (at least, not unless they
	//   are also connected to a NATed or bridged network). These networks are
	//   still useful -- you can join a consul or nomad cluster, SSH between
	//   VMs, serve APIs, run tests, etc.
	//
	// - DHCP stands for "dynamic host configuration protocol". Among other
	//   things it is used to dynamically allocate IP addresses in a local
	//   network. By default, VMware is configured to use DHCP on all non-
	//   bridged networks. Bridged devices are assigned IPs by your local
	//   network's DHCP (usually your router).
	//
	// - Static IPs may also be assigned by way of custom configuration. When a
	//   VM with a static IP is started, VMware will attempt to bind that IP to
	//   the virtual device. DHCP is not used (typically these are configured
	//   directly in the VMX file). However, the bind may fail if another device
	//   on the network is already using the same IP so these are slightly more
	//   complicated.
	//
	// - MAC addresses are a hardware serial number used to identify each device
	//   on the network. Each virtual device, assigned in the VMX file, should
	//   have a unique MAC address.
	//
	// VMware VMs will typically maintain the same IP via DHCP as long as the
	// VM exists (the lease is freed when the VM is destroyed) but the IP may
	// change during host reboots, or if the host networking changes (e.g. when
	// changing wifi networks).
	//
	// VMware Workstation may have up to 256 virtual networks, numbered 0-255
	// and named e.g. vmnet0, vmnet1, etc. By default vmet0 is bridged, vmnet1
	// is host-only, and vmnet8 is NATed. These are configured in
	// /etc/vmware/networking. Each network has several lines of config,
	// including
	//
	//   answer VNET_8_DHCP yes
	//
	// if DHCP is configured. We want to identify all networks where DHCP is
	// configured, and then we can look at the corresponding DHCP lease table in
	// /etc/vmware/vmnet8/dhcpd/dhcpd.leases. Note that a VM may be configured
	// with more than one virtual interface, and in that case there may be more
	// than one entry for that VM.
	//
	// As the notes in dhcpd.leases indicate, each lease is held for a short
	// period of time, with an expiry window specified in UTC. In my experience
	// a VM's IP will not change while the VM is running, unless the host
	// changes networks (e.g. laptop on wifi), so we should only need to worry
	// about finding the current active lease.
	//
	// We can identify which IP corresponds to which VM and which interface(s)
	// by comparing the MAC addresses specified in the vmx file to those found
	// in the lease table.
	//
	// So, after that very long explanation, here's what we're going to do:
	//
	// 1. Inspect the vmx file for any mac addresses and network interfaces.
	// 2. Identify any static IP configuration in the vmx file
	// 3. Inspect /etc/vmware/networking to see which networks are configured
	//    with DHCP
	// 4. Inspect each /etc/vmware/vmnet*/dhcpd/dhcpd.leases to find active
	//    leases that match our mac address(es)
	//
	// Depending on the particulars, we may not need to do all four steps. For
	// example, if a VM has only one IP and it's statically defined, we don't
	// need to query DHCP.
	//
	// It's also technically possible to assign a static IP to a VM by editing
	// the vmnet DHCP configuration in e.g. /etc/vmware/vmnet8/dhcpd/dhcpd.conf.
	// In my experience this is not a common way to do things, so lovm does not
	// currently support it.
	//
	// References:
	// - ps aux | grep vmnet
	// - https://www.vmware.com/support/ws55/doc/ws_net_advanced_ipaddress.html

	// TODO (currently unhandled cases)
	//  - handle static IPs -- ReadIPAddressesFromVMX
	//  - handle bridged networks -- these are assigned by the local network,
	//    not vmnet DHCP, but there may be a way to correlate the virtual
	//    ethernet device with a host address in ifconfig / ip addr
	//  - handle multiple IPs -- ...?
	macs, err := ReadMACAdressesFromVMX(v.VM.Path)
	if err != nil {
		return nil, err
	}
	for _, mac := range macs {
		ip, err := DetectIPFromMACAddress(mac)
		switch err {
		case nil:
			// We found an ip. We're returning now, though there could be others
			return ip, nil
		case ErrNotFound:
			// Could not find an ip address for this MAC, try the next one
			continue
		default:
			// Something weird happened, return to the user
			return nil, err
		}
	}

	return nil, ErrNotFound
}

func (v *VMware) Delete() error {
	// If there's no VM we don't need to do anything
	if !v.Found() {
		return nil
	}

	if err := v.Stop(); err != nil {
		return err
	}

	cmd := exec.Command("vmrun", "deleteVM", v.VM.Path)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("%s", out)
	}

	// Remove the machine path because we don't have a VM anymore
	if err == nil {
		v.VM.Path = ""
	}

	return err
}

// TODO implement Mount
func (v *VMware) Mount() error {
	// TODO check guest tools status because shared folders don't work without
	//  those
	return errNotImplemented
}


func ReadMACAdressesFromVMX(path string) (map[string]net.HardwareAddr, error) {
	macs := map[string]net.HardwareAddr{}
	// example from vmx file:
	// ethernet0.generatedAddress = "00:0c:29:05:6f:e3"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return macs, err
	}

	matches := reGeneratedAddress.FindAllStringSubmatch(string(data), -1)

	for _, match := range matches {
		mac, err := net.ParseMAC(match[2])
		if err != nil {
			log.Printf("error parsing mac address %q from %q: %s\n", match[2], path, err)
			continue
		}
		macs[match[1]] = mac
	}

	return macs, nil
}

func ListDHCPVirtualNetworks() ([]int, error) {
	// example networking config
	//
	// $ cat /etc/vmware/networking
	//
	// VERSION=1,0
	// answer VNET_1_DHCP yes
	// ...
	// answer VNET_8_DHCP yes
	// ...
	var networks []int

	data, err := ioutil.ReadFile(NetworkConfigFile)
	if err != nil {
		return networks, err
	}

	matches := reNetworkingConfig.FindAllStringSubmatch(string(data), -1)

	for _, match := range matches {
		netID, err := strconv.ParseInt(match[1], 10, 0)
		if err != nil {
			log.Printf("Unexpected number format %q", match[1])
			continue
		}
		networks = append(networks, int(netID))
	}

	return networks, nil
}

func FindCurrentLeaseByMAC(netID int, addr net.HardwareAddr) (net.IP, error) {
	// example DHCP lease file
	//
	// $ cat /etc/vmware/vmnet8/dhcpd/dhcpd.leases
	//
	// lease 172.16.23.128 {
	//   starts 2 2019/04/02 01:05:48;
	//   ends 2 2019/04/02 01:35:48;
	//   hardware ethernet 00:0c:29:56:7f:63;
	//   uid ff:bc:9a:4a:2d:00:02:00:00:ab:11:15:39:5e:d3:35:a2:c9:00;
	//   client-hostname "ubuntu";
	// }
	//
	// Note: file is indented using tabs, not spaces as above
	data, err := ioutil.ReadFile(fmt.Sprintf(DHCPLeasesFile, netID))
	if err != nil {
		return nil, err
	}

	matches := reDHCPLeases.FindAllStringSubmatch(string(data), -1)

	for _, match := range matches {
		// Each match should contain a legit entry, but we need to find the one
		// where starts < current time < ends and the MAC address matches.
		// When we find a match we'll return immediately and stop processing.
		MAC, err := net.ParseMAC(match[4])
		if err != nil {
			continue
		}

		// Compare MAC addresses before we spend CPU on anything else.
		if MAC.String() != addr.String() {
			continue
		}

		starts, err := time.Parse(DHCPDateFormat, match[2])
		if err != nil {
			continue
		}

		ends, err := time.Parse(DHCPDateFormat, match[3])
		if err != nil {
			continue
		}

		now := time.Now().UTC()

		// Time matches
		if now.After(starts) && now.Before(ends) {
			// For some reason the API for net.ParseIP does not return an error
			// if there is a problem, so we need to check it ourselves before we
			// return.
			IP := net.ParseIP(match[1])
			if IP != nil {
				return IP, nil
			} else {
				return nil, fmt.Errorf("failed parsing ip address: %q", match[1])
			}
		}
	}

	return nil, ErrNotFound
}

func DetectIPFromMACAddress(mac net.HardwareAddr) (net.IP, error) {
	networks, err := ListDHCPVirtualNetworks()
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		ip, err := FindCurrentLeaseByMAC(network, mac)
		switch err {
		case nil:
			// Yay we found the ip address!
			return ip, nil
		case ErrNotFound:
			// We don't expect to find a match in every network so we'll just
			// continue if we don't find it in this one.
			continue
		default:
			// If we got some other kind of error there was a problem and we'll
			// return control to the user to fix it.
			return nil, err
		}
	}

	return nil, ErrNotFound
}
