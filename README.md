# Updater

CLI´s get widely used. Updating them for various OSes is a pain. The common approach is to add a README.md on the project website an tell the user how to download and install that binary on his local machine.  

Updater will do that for the user with a simple command.

It will download the binary from github releases.

Example usage:

```go
package main

import (
    "github.com/metal-stack/updater"
    "github.com/spf13/cobra"
)

const (
    programName = "mybinary"
    owner = "my-github-organisation"
    repo = "my-cli"
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
            u, err := updater.New(owner, repo, programName)
            if err != nil {
                return err
            }
            return u.Check()
        },
    }
    updateDoCmd = &cobra.Command{
        Use:   "do",
        Short: "do the update of the program",
        RunE: func(cmd *cobra.Command, args []string) error {
            u, err := updater.New(owner, repo, programName)
            if err != nil {
                return err
            }
            return u.Do()
        },
    }

)
```
