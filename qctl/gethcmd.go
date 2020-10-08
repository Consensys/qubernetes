package main

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	//consoleFlags = []cli.Flag{utils.JSpathFlag, utils.ExecFlag, utils.PreloadJSFlag}

	// $> qctl geth attach node1
	gethAttachCommand = cli.Command{
		Name:      "attach",
		Usage:     "attach to the geth process running on a particular node",
		ArgsUsage: "[pod_substring]",
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 1 {
				c.App.Run([]string{"geth", "help", "attach"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()
			namespace := c.String("namespace")
			podName := podNameFromPrefix(nodeName, namespace)
			log.Printf("attaching to geth  pods [%v]", podName)
			cmd := exec.Command("kubectl", "--namespace="+namespace, "exec", "-it", podName, "-c", "quorum", "--", "/geth-helpers/geth-attach.sh")
			dropIntoCmd(cmd)
			fmt.Printf("Bye hope you had fun")
			return nil
		},
	}

	// $> qctl geth exec node1 "eth.blockNumber"
	gethExecCommand = cli.Command{
		Name:      "exec",
		Usage:     "exec a geth command on a particular node",
		ArgsUsage: "[pod_substring] [geth command]",
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 2 {
				c.App.Run([]string{"geth", "help", "exec"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()
			gethCmd := c.Args().Get(1)
			namespace := c.String("namespace")
			podName := podNameFromPrefix(nodeName, namespace)
			log.Printf("executing geth command on pod [%v]", podName)
			cmd := exec.Command("kubectl", "--namespace="+namespace, "exec", "-it", podName, "-c", "quorum", "--", "/geth-helpers/geth-exec.sh", gethCmd)
			dropIntoCmd(cmd)
			return nil
		},
	}
)
