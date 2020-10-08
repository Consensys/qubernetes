package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

// https://github.com/urfave/cli/blob/master/docs/v2/manual.md#getting-started
func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "qctl"
	app.Usage = "command line tool for managing qubernetes network. Yay!"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "namespace, n",
			Aliases: []string{"n", "ns"},
			Value:   "default",
			Usage:   "The k8s namespace for the quorum network",
			EnvVars: []string{"QUORUM_NAMESPACE"},
		},
		//&cli.StringFlag{
		//	Name:  "config, c",
		//	Usage: "Load configuration from (full path) `FILE`",
		//	EnvVars:     []string{"QUBE_CONFIG"},
		//},
	}
	app.Commands = []*cli.Command{

		&logCommand,
		//TODO: should init not take a parameter? what else would you init besides a config?
		&initConfigCommand,
		{
			Name:  "test",
			Usage: "run tests against the running network",
			Subcommands: []*cli.Command{
				&testContractCmd,
				&acceptanceTestRunCmd,
			},
		},
		{
			Name:  "generate",
			Usage: "options for generating base config / resources",
			Subcommands: []*cli.Command{
				&generateNetworkCommand,
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"destroy"},
			Usage:   "options for deleting networks / resources",
			Subcommands: []*cli.Command{
				&networkDeleteCommand,
				// TODO: Think this through a bit, hard delete vs soft delete, etc.
				&nodeDeleteCommand,
			},
		},
		{
			Name:  "stop",
			Usage: "options for stopping nodes.",
			Subcommands: []*cli.Command{
				&nodeStopCommand,
			},
		},
		{
			Name:  "update",
			Usage: "options for updating nodes / resources",
			Subcommands: []*cli.Command{
				&nodeUpdateCommand,
			},
		},
		{
			Name:  "deploy",
			Usage: "options for deploying networks / resources to K8s",
			Subcommands: []*cli.Command{
				&networkCreateCommand,
			},
		},

		{
			Name:  "geth",
			Usage: "options for interacting with geth",
			Subcommands: []*cli.Command{
				&gethAttachCommand,
				&gethExecCommand,
			},
		},

		{
			Name:    "list",
			Aliases: []string{"ls", "get"},
			Usage:   "options for listing resources",
			Subcommands: []*cli.Command{
				&nodeListCommand,
				&allListCommand,
				&urlGetCommand,
				&describeConfigCommand,
				&nodeStatusCommand,
			},
		},

		{
			Name:  "add",
			Usage: "options for adding resources",
			Subcommands: []*cli.Command{
				&nodeAddCommand,
			},
		},

		//{
		//	Name:    "describe",
		//	Aliases: []string{},
		//	Usage:   "options for describing resources",
		//	Subcommands: []*cli.Command{
		//		&nodeListCommand,
		//		&allListCommand,
		//		&urlGetCommand,
		//	},
		//},

		&nodeConnectCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
