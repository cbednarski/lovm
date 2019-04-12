# LOVM

**Lo**ve your **v**irtual **m**achines. Or **lo**w key **v**m **m**anager, or
**lo**cal **vm**. `lovm` is a minimalist, idempotent command-line tool for
managing **local** virtual machines (i.e. VMs you run on your computer).

## Why?

I use regularly use VMs during development to test deployment scripts and
configuration, and to install things I don't want to commingle with my desktop
environment.

I used to use Vagrant but I found that I spent enormous amounts of time building
and maintaining Vagrant boxes; ultimately, not the best use of my time.

I want to satisfy these **minimum** requirements for day-to-day use:

- [x] Clone a VM (installing the OS is outside the scope of `lovm`, we'll just
  clone one that works already)
- [ ] Share a folder with the VM
- [x] Start the VM
- [x] Discover the VM's IP address
- [x] SSH to the VM
- [x] Stop the VM
- [x] Delete the VM when I'm done using it, or when I need to wipe out my changes

## Design Goals

- Simple
- Speedy
- Does not get in the way

## Installing

With [go](https://golang.org/dl/):

    go install github.com/cbednarski/lovm

## Commands

    lovm clone <source>                   Clone a VM. Start here!
    lovm start                            Start the VM
    lovm stop                             Stop the VM
    lovm restart                          Stop and then start the VM
    lovm ssh                              Open an SSH session to the VM
    lovm ip                               Write the VM's IP address to stdout
    lovm mount <host path> <guest path>   Mount a host folder into the VM
    lovm delete                           Delete the VM; get your space back

## Questions

> How do I clone a snapshot?

You can clone a snapshot by adding `:` and the snapshot name to the clone
source. For example:

    lovm clone /path/to/vm:snapshot-name

> How do I ssh to my box?

`lovm ssh` calls `ssh` with the IP address returned by `lovm ip`, but will pass
through any additional flags directly to the underlying `ssh` command. The SSH
prompt is interactive, so you can also type in your password. For example:

    lovm ssh -l root -i ~/.ssh/my_key.pem

Since you can't use `user@ip` syntax to change the ssh login, use `-l` instead.

> Do I have to use VMware Workstation Pro or Fusion Pro?

Yes. `lovm` uses linked clones, which use copy-on-write to make cloning
extremely fast. The non-Pro versions of these products do not (at time of
writing) support linked clones.

> How do I create my own VM?

The easiest way is to use the GUI for VMware or VirtualBox. VMware even has an
"easy install" option that will automatically install popular operating systems
for you. You can also use a VM created using Packer, or provided by someone
else. You can even use a Vagrant Box, though you will need to extract it first.

In order to clone the VM, the source VM must be in a powered-off state (not
suspended), or must have a powered-off snaphot. Since the original VM is non-
destructively cloned each time you run it, you only need to create it once.

> What about all the other virtualization tools, like bhyve and kvm?

I don't use those. I only use VMware, and most people I know only use
VirtualBox.

> What about Windows?

`lovm` doesn't currently run on Windows, though it can run Windows VMs on Linux.
I don't currently develop on Windows, but I know a lot of people do, so I may
add Windows host support later.

> What about snapshots, multiple VMs, or \[feature X\], or \[platform Y\]?

There are a lot of features and platforms I don't use and don't have time to
support. Feel free to clone or fork the project for your own needs. :)

## Comparisons to Other Tools

### Vagrant

Vagrant has many more features like Vagrantfiles (ruby), packaging (Vagrant
Boxes), upload / download of boxes, provisioning, plugins, triggers, remote VM
access (sharing), management of multiple machines, and more. Vagrant has
numerous workflows for working with teams, mass-distribution of VMs, and support
for custom behavior. Vagrant supports VMware through a paid plugin.

`lovm` only supports SSH and basic VM controls. However, it works with VMs that
you already have and does not require extra packaging. VMware is supported out
of the box.

<https://vagrantup.com>

### Docker Desktop

Docker Desktop runs a Linux VM and installs some host-native client tools (like
the Docker CLI) to make docker workflows easy. Docker desktop does not support
other guest operating systems and does not support power management workflows
(instead it runs the VM like an application).

`lovm` manages VMs and does not have any docker-specific features.

### `vmrun`

`lovm` wraps `vmrun`, the CLI included with VMware Workstation. `lovm` saves you
having to re-type vmx paths each time, and also quickly [finds the IP
address][1] when you start a VM.

[1]: https://www.vmware.com/support/ws55/doc/ws_net_advanced_ipaddress.html

## Contributing

I built this for my own use, and I only intend to maintain these features for
myself. Please do feel free to open bugs or PRs or ask questions, but don't feel
bad if I don't respond or don't merge your PR. It's not you. It's me! :)

## Compatibility Matrix

This is not a roadmap.

### Operating Systems

- [ ] FreeBSD
- [x] Linux
- [ ] MacOS
- [ ] Windows

### Virtualization Engines

- [ ] bhyve
- [ ] Docker
- [ ] HyperV
- [ ] Parallels
- [ ] qemu/kvm
- [ ] VirtualBox
- [x] VMware
- [ ] xhyve
