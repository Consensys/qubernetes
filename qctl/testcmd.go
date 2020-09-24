package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os/exec"
)

var (
	// qctl test contract --private node1
	testContractCmd = cli.Command{
		Name:  "contract",
		Usage: "deploy the test contract(s) on a node.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "both",
				Usage: "try to deploy a public and private test contract to the node.",
				Value: true,
			},
			&cli.BoolFlag{
				Name:    "private",
				Aliases: []string{"priv"},
				Usage:   "try to deploy a private test contract to the node.",
			},
			&cli.BoolFlag{
				Name:    "public",
				Aliases: []string{"pub"},
				Usage:   "try to deploy a public test contract to the node.",
			},
		},
		Action: func(c *cli.Context) error {

			if c.Args().Len() < 1 {
				c.App.Run([]string{"test", "help", "contract"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()
			namespace := ""
			podName := podNameFromPrefix(nodeName, namespace)
			fmt.Println(fmt.Sprintf("running test contract(s) on node [%s] pod [%s]", nodeName, podName))

			isBothTest := c.Bool("both")
			isPrivTest := c.Bool("private")
			isPublicTest := c.Bool("public")

			// if either of the specific priv or pub flags are set, only run the specified flags.
			if isPrivTest || isPublicTest {
				isBothTest = false
			}
			var cmd *exec.Cmd
			// test private tx
			if isPrivTest || isBothTest {
				fmt.Println()
				green.Println("  Trying to deploy the test private contract...")
				cmd = exec.Command("kubectl", "--namespace="+namespace, "exec", "-it", podName, "-c", "quorum", "--", "/etc/quorum/qdata/contracts/runscript.sh", "/etc/quorum/qdata/contracts/private_contract.js")
				dropIntoCmd(cmd)
			}

			// test public tx
			if isPublicTest || isBothTest {
				fmt.Println()
				green.Println("  Trying to deploy the test public contract...")
				cmd = exec.Command("kubectl", "--namespace="+namespace, "exec", "-it", podName, "-c", "quorum", "--", "/etc/quorum/qdata/contracts/runscript.sh", "/etc/quorum/qdata/contracts/public_contract.js")
				dropIntoCmd(cmd)
			}

			return nil
		},
	}

	// TODO: support acceptance tests
)
