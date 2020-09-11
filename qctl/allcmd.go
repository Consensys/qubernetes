package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os/exec"
)

var (
	allListCommand = cli.Command{
		Name:    "all",
		Usage:   "list all quorum k8s resources",
		Aliases: []string{"n", "nodes"},
		Action: func(c *cli.Context) error {
			fmt.Println("getting nodes: ", c.Args().First())
			namespace := c.String("namespace")
			cmd := exec.Command("kubectl", "--namespace="+namespace, "get", "all")

			var out bytes.Buffer
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(out.String())
			return nil
		},
	}
)
