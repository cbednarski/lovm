# Using lovm with VirtualBox

By default VirtualBox VMs have an outbound-only NAT network. This means that
inbound connections, like SSH, will not work without additional configuration.
Because the NAT network does not allow any inbound traffic, lovm ip will ignore
this interface when detecting an IP address for the machine.

You have two options to enable SSH.

One is to add VM-specific port forwarding from the host to the guest. For
example, you may forward the host port 40022 to port 22 in the VM. This requires
ad hoc configuration for each VM so lovm ssh does not currently support this.
Additionally, this approach requires a port forward for every protocol (i.e.
to access the VM via http, you have to forward a different port).

The second, more general option is to setup a host-only network. This approach
involves more steps, but will work for any number of VMs and ports, and also
allows lovm ip to detect the machine's ip and allows lovm ssh to connect to the
machine.

## Setting up a host-only network

After a fresh installation of VirtualBox, there are four steps to enable host-
only networking.

1. Create a host-only network (see File -> Host Network Manager)
2. Add a DHCP server to the host-only network
3. Add a second network interface to your source VM (the one you will clone)
4. Make sure the guest operating system is configured to use DHCP for the
   second network interface

You can easily accomplish this via the VirtualBox UI by using the Host Network
Manager to add the host-only network and DHCP server, and then by adding a
second network adapter to the VM. You can also adapt the following CLI commands
to your particular installation:

    vboxmanage hostonlyif create
    vboxmanage dhcpserver add --ifname vboxnet0 --ip 192.168.56.1 \
                              --lowerip 192.168.56.3 --upperip 192.168.56.255 \
                              --netmask 255.255.255.0
    vboxmanage modifyvm /path/to/vm.vbox --hostonlyadapter2 vboxnet0

## Other Networking Situations

VirtualBox supports many other types of networking scenarios. lovm ip, which
underlies the lovm ssh command, can only detect the IP from a host-only network
adapter. It does not work with port forwarding. To learn more about VirtualBox
networking configuration please refer to the manual:

    <https://www.virtualbox.org/manual/ch06.html>
