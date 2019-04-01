# Notes

These are some of the error / exit messages from `vmrun`:

trying to stop VM that is already stopped
```
$ vmrun stop .lovm/epic/epic.vmx
Error: The virtual machine is not powered on: /home/cbednarski/code/epic/.lovm/epic/epic.vmx
```

trying to delete VM that is running
```
$ vmrun deleteVM .lovm/epic/epic.vmx
Error: The virtual machine should not be powered on. It is already running.
```