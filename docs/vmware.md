# Using lovm with VMware

## VMware Pro Features

lovm relies on linked clones for performance. This is a feature of VMware Fusion
Pro (on MacOS) and VMware Workstation Pro (on Windows and Linux) ONLY.

VMware Workstation Player, VMware Fusion, and ESXi are not supported.

You can learn more about these products on the web.

- [VMware Workstation Pro](https://www.vmware.com/products/workstation-pro.html)
- [VMware Fusion / Pro](https://www.vmware.com/products/fusion.html)

## Cloning VMs

VMware can only create a linked clone from a machine that is powered off or from
a powered-off snapshot. lovm will attempt to create a powered-off snapshot for
the source VM before cloning it.

## Snapshots

A snapshot is automatically created during when a source VM is first cloned, if
it does not already exist. If you make changes to the source VM and want those
changes to propagate to clones you make with lovm, you will either need to
update or delete the "lovm" snapshot created for your source VM. If you create
a snapshot with this name, it will be used in any case where the snapshot is
not manually specified in the clone command.
