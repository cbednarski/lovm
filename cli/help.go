package cli

const CommandList = `Commands

  lovm clone <source>                   Clone a VM. Start here!
  lovm start                            Start the VM
  lovm stop                             Stop the VM
  lovm restart                          Stop / start the VM
  lovm ssh                              Open an SSH session to the VM
  lovm ip                               Write the VM's IP address to stdout
  lovm mount <host path> <guest path>   Mount a host folder into the VM
  lovm delete                           Delete the VM
  lovm help                             Show help
`

const ProgramHelp = `LOVM

  A minimalist, idempotent command-line tool for managing local virtual machines

` + CommandList + `
Misc

  Copyright: 2019 Chris Bednarski
  License: MIT
  Contact: https://github.com/cbednarski/lovm
`
