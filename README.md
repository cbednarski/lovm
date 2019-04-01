# LOVM

**Lo**ve your **v**irtual **m**achines. Or **l**ow key **vm** **m**anager, if
you prefer. `lovm` is a minimalist command-line utility for managing local VMs
running in VMware. My goal is to satisfy the **minimum** requirements for
day-to-day use.

`lovm` controls the machine, not the OS. "Stop" means **stop now**, not "Please
kindly shut down when you feel like it."

`lovm` is *extremely* fast and lightweight. How fast?

```
cbednarski@stormbuntu:~/code/example$ \
    date && \
    lovm clone ~/vmware/Ubuntu\ 64-bit/Ubuntu\ 64-bit.vmx && \
    lovm start
Sat Mar 30 21:33:39 PDT 2019
2019/03/30 21:33:43 machine "/home/cbednarski/code/example/.lovm/example/example.vmx" running (vmware)
```

That's about **4 seconds to clone and start a VM.**

## Goals

- [x] I want a linked clone, not a full clone, to optimize for speed and storage
- [ ] I want shared folders to be re-enabled each time I restart the VM
- [ ] I want to know what the IP address of the VM is so I can SSH to it
- [ ] I want things to *just work* so I don't have to waste time retyping
      commands or figuring out workarounds
- [ ] Simple is better

## Commands

    lovm clone <source>                   Clone a VM. Start here!
    lovm start                            Start the VM
    lovm stop                             Stop the VM
    lovm restart                          Restart, or poweroff/poweron the VM
    lovm ssh                              Open an SSH session to the VM
    lovm ip                               Write the VM's IP address to stdout
    lovm mount <host path> <guest path>   Mount a host folder into the VM
    lovm delete                           Delete the VM; get your space back

## Compatibility Matrix

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

**Note:** Linked clones are only supported on VMware Workstation Pro and VMware
Fusion Pro, so this tool will not work with Fusion / Player / ESXi.

If you are looking for a virtual machine tool that has more features and better
support for other operating systems and virtualization platforms, check out
[Vagrant](https://vagrantup.com) which supports all of the above platforms.
