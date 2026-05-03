package kill

import (
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:  "kill",
	Usage: "Kill all cactus-* processes",
	Action: func(_ *cli.Context) error {
		pterm.Info.Println("Killing all cactus-* processes...")
		if err := killProcesses(); err != nil {
			pterm.Warning.Printfln("Kill returned error: %v", err)
		} else {
			pterm.Success.Println("Done")
		}
		return nil
	},
}
