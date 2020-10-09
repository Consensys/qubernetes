package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
			&cli.StringFlag{
				Name:    "config",
				Usage:   "the path of the qubernetes config file being used.",
				EnvVars: []string{"QUBE_CONFIG"},
			},
			&cli.BoolFlag{
				Name:    "wait",
				Usage:   "wait for the network to become available, e.g. all pods running, before exiting.",
				Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			wait := c.Bool("wait")
			var configFileYaml QConfig
			if wait { // if the wait flag is set, the qubernetes flag / env var is required, to check the status of the network.
				config := c.String("config")
				configYaml, err := LoadYamlConfig(config)
				if err != nil || config == "" {
					log.Fatal("config file [%v] could not be loaded into the valid qubernetes yaml. err: [%v]", config, err)
				}
				configFileYaml = configYaml
			}
			err := k8sCreateDeleteCluster(c, "apply")
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error while trying to create k8s cluster [%v]", err), 3)
			}
			if wait {
				waitForPodsReadyState(configFileYaml)
			}
			return nil
		},
	}
	nodeStatusCommand = cli.Command{
		Name:  "status",
		Usage: "list the status of the running network.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "k8s-dir",
				Usage:    "the path of the dir containing the K8s resource yaml.",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			config := c.String("config")
			configFileYaml, err := LoadYamlConfig(config)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid qubernetes yaml. err: [%v]", config, err)
			}
			waitForPodsReadyState(configFileYaml)
			return nil
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
				Usage: "create or re-create all config, this is a destructive op.",
			},
			&cli.StringFlag{ // this is only required to get the enodeurl
				Name:    "k8sdir",
				Usage:   "The k8sdir (usually out) containing the output k8s resources",
				EnvVars: []string{"QUBE_K8S_DIR"},
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
			k8sdir := c.String("k8sdir")

			// If we are running an update, the k8s directory is required, because the additional resources will be generate there.
			if k8sdir == "" && update {
				c.App.Run([]string{"qctl", "help", "init"})
				red.Println("When updating a network, k8sdir must be provided! Either via he flag [k8sdir] or env var [QUBE_K8S_DIR]")
				return cli.Exit(fmt.Sprintf("When updating a network, k8sOutDir must be provided!  k8sOutDir [%s]", k8sdir), 3)
			} else if k8sdir == "" { // if the k8sdir is not set and it is not an update, set it to the default out directoy in the current dir.
				k8sdir = pwd + "/out"
			}

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
			if k8sdir != "" {
				green.Println("  Using k8soutdir:")
				fmt.Println()
				fmt.Println("  " + k8sdir)
				fmt.Println()
			}
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
				log.Fatal("config file [%v] could not be loaded into the valid qubernetes yaml. err: [%v]", configFile, err)
			}

			// if the quberentes version is set to latest, try to pull it from the remote, as it may have changed upstream.
			if qubernetesVersion == "latest" {
				fmt.Println("trying to pull latest container")
				pullContainerCmd := exec.Command("docker", "pull", "quorumengineering/qubernetes:latest")
				err = dropIntoCmd(pullContainerCmd)
				if err != nil {
					fmt.Println()
					red.Println(fmt.Sprintf("Error running trying to generate network resources."))
					red.Println(fmt.Sprintf("Command that failed: [%v]", pullContainerCmd.String()))
					red.Println(fmt.Sprintf("Is Docker running on your machine? [%v]", err))
					fmt.Println()
					return cli.Exit(fmt.Sprintf("Docker must be running on host, cmd failed [%v]", pullContainerCmd.String()), 3)
				}
				fmt.Println()
				fmt.Println()
			}

			cmd := exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", k8sdir+":/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "qubes.yaml")
			if update {
				cmd = exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", k8sdir+":/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "--action=update", "qubes.yaml")
			} else if create {
				cmd = exec.Command("docker", "run", "--rm", "-it", "-v", configFile+":/qubernetes/qubes.yaml", "-v", k8sdir+":/qubernetes/out", "quorumengineering/qubernetes:"+qubernetesVersion, "./qube-init", "--action=create", "qubes.yaml")
			}

			//fmt.Println(cmd)
			err = dropIntoCmd(cmd)
			if err != nil {
				fmt.Println()
				red.Println(fmt.Sprintf("Error running trying to generate network resources with the qubernetes container."))
				red.Println(fmt.Sprintf("Command that failed:"))
				red.Println(fmt.Sprintf(cmd.String()))
				fmt.Println()
				red.Println(fmt.Sprintf("Is Docker running on your machine? [%v]", err))
				fmt.Println()
				return cli.Exit(fmt.Sprintf("Docker must be running on host, cmd failed \n %v", cmd.String()), 3)
			}
			fmt.Println()

			fmt.Println("=======================================================================================")
			fmt.Println()
			green.Println("  The Quorum and K8s resources have been generated in the directory:")
			fmt.Println()
			fmt.Println("  " + k8sdir)
			fmt.Println()
			fmt.Println("  Using config file:")
			fmt.Println()
			fmt.Println("  " + configFile)
			fmt.Println()
			fmt.Println()
			// tell the defaults
			fmt.Println("  Network Configuration: ")
			green.Println(fmt.Sprintf("  num nodes = %d", len(configFileYaml.Nodes)))
			green.Println(fmt.Sprintf("  consensus = %s", configFileYaml.Genesis.Consensus))
			green.Println(fmt.Sprintf("  quorumVersion = %s", configFileYaml.Genesis.QuorumVersion))
			green.Println(fmt.Sprintf("  (node1) txManager = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.Name))
			green.Println(fmt.Sprintf("  (node1) tmVersion = %s", configFileYaml.Nodes[0].QuorumEntry.Tm.TmVersion))
			green.Println(fmt.Sprintf("  (node1) chainId = %s", configFileYaml.Genesis.Chain_Id))
			fmt.Println()
			fmt.Println("  To enable future commands, e.g. qctl create network, qctl delete network, to use this network ")
			fmt.Println("  config, set the QUBE_K8S_DIR environment variable to the out directory that has just been generated")
			fmt.Println("  by running: ")
			fmt.Println()
			fmt.Println("*****************************************************************************************")
			green.Println(fmt.Sprintf("  $> export QUBE_K8S_DIR=%s", k8sdir))
			green.Println(fmt.Sprintf("  $> qctl deploy network --wait"))
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
		return err
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
