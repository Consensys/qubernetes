package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"strings"
)

var (
	networkDeleteCommand = cli.Command{
		Name:  "network",
		Usage: "delete a quorum k8s network given the dir holding the k8s yaml resources.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "k8s-dir",
				Usage:    "the path of the dir containing the K8s resource yaml.",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return k8sCreateDeleteCluster(c, "delete")
		},
	}

	networkCreateCommand = cli.Command{
		Name:  "network",
		Usage: "create a quorum k8s network given the dir holding the k8s yaml resources.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "k8s-dir",
				Usage:    "the path of the dir containing the K8s resource yaml.",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return k8sCreateDeleteCluster(c, "apply")
		},
	}

	// $>  qctl --config-file=qubernetes.yaml
	// qctl init -f
	// qctl init
	// TODO: break out into:
	// qctl generate config
	// qctl init // use env QUBE_CONFIG
	generateNetworkCommand = cli.Command{
		Name:  "network",
		Usage: "creates new resources for both quorum and Kubernetes.",
		//#ArgsUsage: "[pod_substring]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "create",
				Usage: "create or re-create all config, this is a discructive op.",
			},
			&cli.BoolFlag{
				Name:  "update",
				Usage: "update only the config for changed nodes, e.g. add new keys, but don't delete current keys.",
			},
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

			// update / create
			update := c.Bool("update")
			create := c.Bool("create")

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
			// if the flags --update or --create are not set, prompt the user if the config already exists.
			cmd := exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", pwd+"/out:/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "qubes.yaml")
			if update {
				cmd = exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", pwd+"/out:/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "--action=update", "qubes.yaml")
			} else if create {
				cmd = exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", pwd+"/out:/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "--action=create", "qubes.yaml")
			}

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
			green.Println(fmt.Sprintf("  (node1) txManger = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.Name))
			green.Println(fmt.Sprintf("  (node1) tmVersion = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.TmVersion))
			green.Println(fmt.Sprintf("  (node1) chainId = %s", configFileYaml.Genesis.Chain_Id))
			fmt.Println()
			fmt.Println("  To enable future commands, e.g. qctl create network, qctl delete network, to use this network ")
			fmt.Println("  config, set the QUBE_K8S_DIR environment variable to the out directory that has just been generated")
			fmt.Println("  by running: ")
			fmt.Println()
			fmt.Println("*****************************************************************************************")
			green.Println(fmt.Sprintf("  $> export QUBE_K8S_DIR=%s", k8sOutDir))
			green.Println(fmt.Sprintf("  $> qctl deploy network"))
			fmt.Println("*****************************************************************************************")
			fmt.Println()
			return nil
		},
	}
)

func k8sCreateDeleteCluster(c *cli.Context, action string) error {
	k8sdir := c.String("k8s-dir")
	// if the passed in k8s dir does not exit, tell the user and do not proceed.
	if _, err := os.Stat(k8sdir); os.IsNotExist(err) {
		log.Error("the --k8s-dir [%v] does not exist!", k8sdir)
	}
	namespace := c.String("namespace")
	log.Printf("%s network in k8sdir [%v]", action, k8sdir)

	var cmd *exec.Cmd
	if _, err := os.Stat(k8sdir + "/deployments"); os.IsNotExist(err) {
		cmd = exec.Command("kubectl", "--namespace="+namespace, action, "-f", k8sdir)
	} else {
		cmd = exec.Command("kubectl", "--namespace="+namespace, action, "-f", k8sdir, "-f", k8sdir+"/deployments")
	}
	fmt.Println(cmd.String())
	dropIntoCmd(cmd)
	return nil
}
