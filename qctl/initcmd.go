package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os/exec"
	"strings"
)

var (
	// $>  qctl --config-file=qubernetes.yaml
	// qctl init -f
	// qctl init
	// TODO: break out into:
	// qctl generate config
	// qctl init // use env QUBE_CONFIG
	initCommand = cli.Command{
		Name:  "generate",
		Usage: "creates new resources for both quorum and Kubernetes.",
		//#ArgsUsage: "[pod_substring]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FULL_PATH_FILE`",
				EnvVars: []string{"QUBE_CONFIG"},
				//Required: true,
			},
			&cli.StringFlag{
				Name:  "version",
				Usage: "Which version of qubernetes to use.",
				Value: "latest",
				//Required: true,
			},
			&cli.BoolFlag{
				Name:    "force",
				Usage:   "Initialize new network, if existing out folder exists, delete it without prompting.",
				Aliases: []string{"f"},
			},
		},

		Action: func(c *cli.Context) error {
			qubernetesVersion := c.String("version")

			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

			//force := c.Bool("force")
			configFile := c.String("config")

			if configFile == "" {
				c.App.Run([]string{"qctl", "help", "init"})

				// QUBE_CONFIG or flag
				fmt.Println()

				fmt.Println()
				red.Println("  --config flag must be provided.")
				red.Println("             or ")
				red.Println("     QUBE_CONFIG environment variable needs to be set to your config file.")
				fmt.Println()
				red.Println(" If you need to generate a qubernetes.yaml config use the command: ")
				fmt.Println()
				green.Println("   qctl generate config")
				fmt.Println()
				return cli.Exit("--config flag must be set to the fullpath of your config file.", 3)
			}
			fmt.Println()
			green.Println("  Using config file:")
			fmt.Println()
			fmt.Println("  " + configFile)
			fmt.Println()
			fmt.Println("*****************************************************************************************")
			fmt.Println()
			// the config file must exist or this is an error.
			if fileExists(configFile) {
				// check if config file is full path or relative path.
				if !strings.HasPrefix(configFile, "/") {
					configFile = pwd + "/" + configFile
				}

			} else {
				c.App.Run([]string{"qctl", "help", "init"})
				return cli.Exit(fmt.Sprintf("ConfigFile must exist! Given configFile [%v]", configFile), 3)
			}
			configFileYaml, err := LoadYamlConfig(configFile)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			//log.Printf("create network from configfile [%v] pwd [%v]", configFile, pwd)
			//docker run --rm -it -v /Users/libby/Workspace.Quorum/qubernetes-priv/qctl/:/qubernetes/qubes.yaml -v /Users/libby/Workspace.Quorum/qubernetes-priv/qctl/out:/qubernetes/out quorumengineering/qubernetes:0.1.0.2 ./qube-init qubes.yaml
			//cmd := exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", pwd+"/out:/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "qubes.yaml")
			//if force {
			cmd := exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", pwd+"/out:/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "--action=create", "qubes.yaml")
			//}

			//fmt.Println(cmd)
			dropIntoCmd(cmd)
			fmt.Println()

			k8sOutDir := pwd + "/out"
			fmt.Println("=======================================================================================")
			fmt.Println()
			green.Println("  The Quorum and K8s resources have been generated in the directory:")
			fmt.Println()
			fmt.Println("  " + k8sOutDir)
			fmt.Println()
			fmt.Println("  Using config file:")
			fmt.Println()
			fmt.Println("  " + configFile)
			fmt.Println()
			//fmt.Println("The Quorum network values are:")
			fmt.Println()
			// tell the defaults
			fmt.Println("  Network Configuration: ")
			green.Println(fmt.Sprintf("  num nodes = %d", len(configFileYaml.Nodes)))
			green.Println(fmt.Sprintf("  consensus = %s", configFileYaml.Genesis.Consensus))
			green.Println(fmt.Sprintf("  quorumVersion = %s", configFileYaml.Genesis.QuorumVersion))
			green.Println(fmt.Sprintf("  (node1) transacationManger = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.Name))
			green.Println(fmt.Sprintf("  (node1) tmVersion = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.TmVersion))
			green.Println(fmt.Sprintf("  (node1) chainId = %s", configFileYaml.Genesis.Chain_Id))
			fmt.Println()
			fmt.Println("  To enable future commands, e.g. qctl create network, qctl delete network, to use this network ")
			fmt.Println("  config, set the QUBE_K8S_DIR environment variable to the out directory that has just been generated")
			fmt.Println("  by running: ")
			fmt.Println()
			fmt.Println("*****************************************************************************************")
			green.Println(fmt.Sprintf("  export QUBE_K8S_DIR=%s", k8sOutDir))
			fmt.Println("*****************************************************************************************")
			fmt.Println()
			return nil
		},
	}
)
