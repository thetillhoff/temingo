/*
Copyright © 2023 Till Hoffmann <till@thetillhoff.de>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var version = "dev" // This is just the default. The actual value is injected at compiletime

// versionCommand represents the version command
var versionCommand = &cli.Command{
	Name:  "version",
	Usage: "Prints the version of temingo",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		fmt.Println(version)
		return nil
	},
}
