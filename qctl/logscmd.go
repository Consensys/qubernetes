package main

import (
	"fmt"
	//log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os/exec"
)

var (
	//consoleFlags = []cli.Flag{utils.JSpathFlag, utils.ExecFlag, utils.PreloadJSFlag}
	// $> qctl --ns=quorum-test logs --follow 1 quorum
	// $> qctl logs --follow 1 quorum
	// $> qctl logs node1 quorum
	// $> qctl logs node1 tessera
	logCommand = cli.Command{
		Name:      "log",
		Aliases:   []string{"logs"},
		Usage:     "Show logs for [quorum, tessera, constellation], running on a specific pod",
		ArgsUsage: "[pod_substring] [quorum | tessera | constellation]",
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			fmt.Fprintf(c.App.Writer, "for shame\n")
			return err
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "follow",
				Aliases: []string{"f"},
				Usage:   "Follow logs",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 1 {
				c.App.Run([]string{"qctl", "help", "logs"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()
			// quorum, tessera or constellation
			container := c.Args().Get(1)
			if container == "" {
				container = "quorum"
			}
			follow := c.Bool("follow")
			namespace := c.String("namespace")
			//fmt.Println("namespace", namespace)
			podName := podNameFromPrefix(nodeName, namespace)
			if podName == "" {
				fmt.Println("No Pods found with substring ", nodeName)
				showPods(namespace)
				return cli.Exit(c.App.Command("logs").Usage, 2)
			}
			//  logs -f quorum-node1-deployment-7b6c4c8d8-tkxww quorum
			var cmd *exec.Cmd
			if follow {
				cmd = exec.Command("kubectl", "--namespace="+namespace, "logs", "--follow", podName, container)
			} else {
				cmd = exec.Command("kubectl", "--namespace="+namespace, "logs", podName, container)
			}
			dropIntoCmd(cmd)
			return nil
		},
	}
)
