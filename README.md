# Updater

CLI´s get widely used. Updating them for various OSes is a pain. The common approach is to add a README.md on the project website an tell the user how to download and install that binary on his local machine.  

Updater will do that for the user with a simple command.

Example usage:

```go
package main

import (
    "github.com/metal-stack/updater"
    "github.com/spf13/cobra"
)

const (
    programName = "mybinary"
    downloadURLPrefix = "https://my.website.com/" + programName
)

var (
    updateCmd = &cobra.Command{
        Use:   "update",
        Short: "update the program",
    }
    updateCheckCmd = &cobra.Command{
        Use:   "check",
        Short: "check for update of the program",
        RunE: func(cmd *cobra.Command, args []string) error {
            u := updater.New(downloadURLPrefix, programName)
            return u.Check()
        },
    }
    updateDoCmd = &cobra.Command{
        Use:   "do",
        Short: "do the update of the program",
        RunE: func(cmd *cobra.Command, args []string) error {
            u := updater.New(downloadURLPrefix, programName)
            return u.Do()
        },
    }
    updateDumpCmd = &cobra.Command{
        Use:   "dump <binary>",
        Short: "dump the version update file",
        RunE: func(cmd *cobra.Command, args []string) error {
            u := updater.New(downloadURLPrefix, programName)
            return u.Dump(args[0])
        },
    }
)
```

You need to place the following files on a webserver:

```bash
mybinary-linux-amd64
version-linux-amd64.json
```

Where *mybinary-linux-amd64* is the actual binary. You also can of course create additional binaries for Windows and Darwin accordingly.  

*version-linux-amd64.json* is a JSON file containing version and checksum of the binary. It can be generated with the `dump` command shown above.
