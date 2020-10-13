package main

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	//qctl add cakeshop
	cakeshopAddCommand = cli.Command{
		Name:    "cakeshop",
		Usage:   "add cakeshop to the network",
		Aliases: []string{"cake"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "version",
				Aliases: []string{"qv"},
				Usage:   "Quorum Version.",
				Value:   "latest",
			},
		},
		Action: func(c *cli.Context) error {
			cakeVersion := c.String("version")
			configFile := c.String("config")

			// get the current directory path, we'll use this in case the config file passed in was a relative path.
			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

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
				log.Fatal("config file [%v] could not be loaded into the valid qubernetes yaml. err: [%v]", configFile, err)
			}
			if configFileYaml.Cakeshop.Version == "" {
				configFileYaml.Cakeshop.Version = cakeVersion
			} else {
				green.Println(" cakeshop is already set in the config")
				return nil
			}

			fmt.Println(fmt.Sprintf("Adding cakeshop version [%s]", cakeVersion))
			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			fmt.Println("cakeshop has been added to the config file [%s]", configFile)
			fmt.Println("Next, generate the additional resources for cakeshop on k8s:")
			fmt.Println()
			fmt.Println("**********************************************************************************************")
			fmt.Println()
			green.Println(fmt.Sprintf("  $> qctl generate network --update"))
			fmt.Println()
			fmt.Println("**********************************************************************************************")

			return nil
		},
	}
	//qctl delete cakeshop
	cakeshopDeleteCommand = cli.Command{
		Name:    "cakeshop",
		Usage:   "delete cakeshop from the network",
		Aliases: []string{"cake"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{ // this is only required if the user wants to delete the generated (k8s/quorum) resources as well.
				Name:    "k8sdir",
				Usage:   "The k8sdir (usually out) containing the output k8s resources",
				EnvVars: []string{"QUBE_K8S_DIR"},
			},
		},
		Action: func(c *cli.Context) error {
			configFile := c.String("config")

			k8sdir := c.String("k8sdir")
			// get the current directory path, we'll use this in case the config file passed in was a relative path.
			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

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

			rmDeployment := exec.Command("kubectl", "delete", "-f", k8sdir+"/07-cakeshop.yaml")
			fmt.Println("rmDeployment", rmDeployment)
			runCmd(rmDeployment)

			rmDeploymentFile := exec.Command("rm", "-f", k8sdir+"/07-cakeshop.yaml")
			runCmd(rmDeploymentFile)

			configFileYaml, err := LoadYamlConfig(configFile)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid qubernetes yaml. err: [%v]", configFile, err)
			}
			// remove from yaml
			configFileYaml.Cakeshop = Cakeshop{}
			fmt.Println("Deleting cakeshop and associated resources.")
			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			fmt.Println(fmt.Sprintf("cakeshop has been removed from the config file [%s]", configFile))
			//fmt.Println("Next, generate the resources without cakeshop support cakeshop on k8s:")
			//fmt.Println()
			//fmt.Println("**********************************************************************************************")
			//fmt.Println()
			//green.Println(fmt.Sprintf("  $> qctl generate network --update"))
			fmt.Println()
			fmt.Println("**********************************************************************************************")

			return nil
		},
	}
)
