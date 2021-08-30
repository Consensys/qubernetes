package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// qctl generate config --qversion=21.7.1 --consensus=istanbul --tmversion=21.7.2 --tm=tessera --num=4
	// qctl generate config --num=5 --qversion=21.7.1
	initConfigCommand = cli.Command{
		Name:  "init",
		Usage: "creates a base qubernetes.yaml file which can be used to create a Quorum network.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FULL_PATH_FILE`",
				EnvVars: []string{"QUBE_CONFIG"},
			},
			&cli.IntFlag{
				Name:  "num",
				Usage: "Number of nodes in the network.",
				Value: DefaultNodeNumber,
			},
			&cli.StringFlag{
				Name:  "consensus",
				Usage: "Consensus to use raft | istanbul | clique.",
				Value: DefaultConesensus,
			},
			&cli.StringFlag{
				Name:  "qibftblock",
				Usage: "Blocknumber at which the network will switch over from ibft to qibft.",
				Value: "0",
			},
			&cli.StringFlag{
				Name:    "qversion",
				Aliases: []string{"qv"},
				Usage:   "Quorum Version.",
				Value:   DefaultQuorumVersion,
			},
			&cli.StringFlag{
				Name:    "tmversion",
				Aliases: []string{"tmv"},
				Usage:   "transaction Manager Version.",
				Value:   DefaultTesseraVersion,
			},
			&cli.StringFlag{
				Name:  "tm",
				Usage: "transaction Manager to user: tessera | constellation.",
				Value: DefaultTmName,
			},
			&cli.StringFlag{
				Name:  "chainid",
				Usage: "The chain id for the network.",
				Value: DefaultChainId,
			},
			&cli.StringFlag{
				Name:  "qimagefull",
				Usage: "The full repo + image name of the quorum image.",
			},
			&cli.StringFlag{
				Name:  "tmimagefull",
				Usage: "The full repo + image name of the tm image.",
			},
			&cli.StringFlag{
				Name:  "gethparams",
				Usage: "additional geth startup params to run on the node.",
			},
			&cli.BoolFlag{
				Name:    "monitor",
				Aliases: []string{"monit"},
				Usage:   "enable monitoring on the geth / quorum node (prometheus).",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "cakeshop",
				Aliases: []string{"cake"},
				Usage:   "deploy cakeshop with the Quorum network.",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "ingress",
				Usage: "include a K8s ingress with the Quorum network.",
				Value: false,
			},
		},

		Action: func(c *cli.Context) error {
			pwdCmd := exec.Command("pwd")
			b, _ := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())
			// If the QUBE_CONFIG env is set or the flag passed in, use this file path and generate the config there.
			// this is helpful when creating, deleting, networks repeatedly so that the config dirs can be set once and
			// will be generated to the same place.
			configFile := c.String("config")
			if configFile == "" {
				// no configuration file provided, check for flags and use the default.
				configFile = pwd + "/qubernetes.generate.yaml"
			}

			// TODO: it might be nice to allow these to override the config file, load the config then set any additional
			// params that were passed in.
			numberNodes := c.Int("num")
			quorumVersion := c.String("qversion")
			tmVersion := c.String("tmversion")
			transactionManager := c.String("tm")
			consensus := c.String("consensus")
			qibftblock := c.String("qibftblock")
			chainId := c.String("chainid")
			qimagefull := c.String("qimagefull")
			tmImageFull := c.String("tmimagefull")
			gethparams := c.String("gethparams")
			isMonitoring := c.Bool("monitor")
			isCakeshop := c.Bool("cakeshop")

			// k8s specific
			isIngress := c.Bool("ingress")

			configYaml := QConfig{}
			for i := 1; i <= numberNodes; i++ {
				quorum := Quorum{
					Consensus:      consensus,
					QuorumVersion:  quorumVersion,
					DockerRepoFull: qimagefull,
				}
				tm := Tm{
					Name:           transactionManager,
					TmVersion:      tmVersion,
					DockerRepoFull: tmImageFull,
				}
				quorumEntry := QuorumEntry{
					Quorum: quorum,
					Tm:     tm,
				}
				// for now we need to replace =" with =\"  and *" with *\"
				// this cleans the params for inserting them into the pod, as " must be escaped.
				if gethparams != "" {
					gethparams = strings.ReplaceAll(gethparams, "=\"", "=\\\"")
					gethparams = strings.ReplaceAll(gethparams, "*\"", "*\\\"")
				}
				gethEntry := GethEntry{
					GetStartupParams: gethparams,
				}
				nodeEntry := NodeEntry{
					NodeUserIdent: fmt.Sprintf("quorum-node%d", i),
					KeyDir:        fmt.Sprintf("key%d", i),
					QuorumEntry:   quorumEntry,
					GethEntry:     gethEntry,
				}
				configYaml.Nodes = append(configYaml.Nodes, nodeEntry)

			}

			configYaml.Genesis.QuorumVersion = quorumVersion
			configYaml.Genesis.TmVersion = tmVersion
			configYaml.Genesis.Consensus = consensus
			if consensus == "qibft" { // if --qibftblock=BLOCK_NUM not specified set to default block 0, else the user can specify --qibftblock=BLOCK_NUM
				configYaml.Genesis.QibftBlock = qibftblock
			}
			configYaml.Genesis.Chain_Id = chainId

			if isMonitoring {
				//configYaml.Prometheus.NodePort = DefaultPrometheusNodePort
				configYaml.Prometheus.Enabled = true
			}
			if isCakeshop {
				configYaml.Cakeshop.Version = "latest"
			}

			if isIngress {
				configYaml.K8s.Service.Type = ServiceTypeNodePort
				configYaml.K8s.Service.Ingress.Strategy = "OneToMany"
			}
			configBytes := []byte(configYaml.ToString())

			// try to write the generated file out to disk this file will be used to initialize the network.
			err := ioutil.WriteFile(configFile, configBytes, 0644)
			if err != nil {
				log.Fatal("error writing configFile to [%v]. err: [%v]", configFile, err)
			}
			// TODO: check the config file was properly generated
			// Set the configfile to the full path
			if fileExists(configFile) {
				// check if config file is full path or relative path.
				if !strings.HasPrefix(configFile, "/") {
					configFile = pwd + "/" + configFile
				}
			}
			fmt.Println()

			green.Println("=======================================================================================")
			fmt.Println()
			fmt.Println("Your Qubernetes config has been generated see:")
			fmt.Println()
			fmt.Println("  " + configFile)
			fmt.Println()
			fmt.Println("The Quorum network values are:")
			fmt.Println()
			// tell the default
			green.Println(fmt.Sprintf("  num nodes = %d", numberNodes))
			green.Println(fmt.Sprintf("  consensus = %s", consensus))
			green.Println(fmt.Sprintf("  quorumVersion = %s", quorumVersion))
			green.Println(fmt.Sprintf("  tmVersion = %s", tmVersion))
			green.Println(fmt.Sprintf("  transactionManager = %s", transactionManager))
			green.Println(fmt.Sprintf("  chainId = %s", chainId))
			fmt.Println()
			fmt.Println("To set this as your default config for future commands, run: ")
			fmt.Println()
			fmt.Println("**********************************************************************************************")
			fmt.Println()
			green.Println(fmt.Sprintf("  $> export QUBE_CONFIG=%s", configFile))
			fmt.Println()
			green.Println(fmt.Sprintf("  $> qctl generate network --create"))
			fmt.Println()
			fmt.Println("**********************************************************************************************")
			return nil
		},
	}
	describeConfigCommand = cli.Command{
		Name:  "config",
		Usage: "displays info about the quberentes config.",
		//#ArgsUsage: "[pod_substring]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FULL_PATH_FILE`",
				EnvVars: []string{"QUBE_CONFIG"},
				//Required: true,
			},
			&cli.StringFlag{ // this is only required to get the enodeurl
				Name:    "k8sdir",
				Usage:   "The k8sdir (usually out) containing the output k8s resources",
				EnvVars: []string{"QUBE_K8S_DIR"},
			},
			&cli.BoolFlag{
				Name:  "long, l",
				Usage: "Display all relavent information from the config",
				//Required: true,
			},
		},

		Action: func(c *cli.Context) error {

			pwdCmd := exec.Command("pwd")
			b, _ := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

			configFile := c.String("config")
			outputLong := c.Bool("long")
			k8sdir := c.String("k8sdir")

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
			//TODO: get the global or passed in k8s dir.
			fmt.Println()
			fmt.Println("=======================================================================================")
			fmt.Println()
			green.Println("  Using qubernetes config file:")
			fmt.Println()
			fmt.Println("  " + configFile)
			fmt.Println()
			green.Println("  Using k8sdir:")
			fmt.Println()
			if k8sdir != "" {
				fmt.Println("  " + k8sdir)
			} else {
				fmt.Println("  NOT SET")
			}
			fmt.Println()
			fmt.Println()
			fmt.Println("  To export:")
			fmt.Println()
			green.Println("  export QUBE_CONFIG=" + configFile)
			if k8sdir != "" {
				green.Println("  export QUBE_K8S_DIR=" + k8sdir)
			}
			fmt.Println()
			fmt.Println("=======================================================================================")
			fmt.Println()

			// display the config contents
			fmt.Println("  Network Configuration: ")
			fmt.Println()
			if outputLong {
				displayConfigLong(configFileYaml)
			} else {
				fmt.Println("only display the first node")
				fmt.Println("to display all nodes, run: qctl ls config --long")
				displayConfigShort(configFileYaml)
			}
			fmt.Println()
			return nil
		},
	}
)

// TODO: could be smarter here and only display nodes that differ from eachother, e.g. diff versions of quorum / tessera.
func displayConfigLong(configFileYaml QConfig) {
	green.Println(fmt.Sprintf("  num nodes = %d", len(configFileYaml.Nodes)))
	green.Println(fmt.Sprintf("  consensus = %s", configFileYaml.Genesis.Consensus))
	green.Println(fmt.Sprintf("  quorumVersion = %s", configFileYaml.Genesis.QuorumVersion))
	for i := 0; i < len(configFileYaml.Nodes); i++ {
		fmt.Println()
		green.Println(fmt.Sprintf("     [%s] transactionManager = %s", configFileYaml.Nodes[i].NodeUserIdent, configFileYaml.Nodes[i].QuorumEntry.Tm.Name))
		green.Println(fmt.Sprintf("     [%s] tmVersion = %s", configFileYaml.Nodes[i].NodeUserIdent, configFileYaml.Nodes[i].QuorumEntry.Tm.TmVersion))
		green.Println(fmt.Sprintf("     [%s] quorumVersion = %s", configFileYaml.Nodes[i].NodeUserIdent, configFileYaml.Nodes[i].QuorumEntry.Quorum.QuorumVersion))
		green.Println(fmt.Sprintf("     [%s] consensus = %s", configFileYaml.Nodes[i].NodeUserIdent, configFileYaml.Nodes[i].QuorumEntry.Quorum.Consensus))
		green.Println(fmt.Sprintf("     [%s] chainId = %s", configFileYaml.Nodes[i].NodeUserIdent, configFileYaml.Genesis.Chain_Id))
		fmt.Println()
	}
}

func displayConfigShort(configFileYaml QConfig) {
	green.Println(fmt.Sprintf("  num nodes = %d", len(configFileYaml.Nodes)))
	green.Println(fmt.Sprintf("  consensus = %s", configFileYaml.Genesis.Consensus))
	green.Println(fmt.Sprintf("  quorumVersion = %s", configFileYaml.Genesis.QuorumVersion))
	fmt.Println()
	green.Println(fmt.Sprintf("     [%s] transactionManager = %s", configFileYaml.Nodes[0].NodeUserIdent, configFileYaml.Nodes[0].QuorumEntry.Tm.Name))
	green.Println(fmt.Sprintf("     [%s] tmVersion = %s", configFileYaml.Nodes[0].NodeUserIdent, configFileYaml.Nodes[0].QuorumEntry.Tm.TmVersion))
	green.Println(fmt.Sprintf("     [%s] quorumVersion = %s", configFileYaml.Nodes[0].NodeUserIdent, configFileYaml.Nodes[0].QuorumEntry.Quorum.QuorumVersion))
	green.Println(fmt.Sprintf("     [%s] consensus = %s", configFileYaml.Nodes[0].NodeUserIdent, configFileYaml.Nodes[0].QuorumEntry.Quorum.Consensus))
	green.Println(fmt.Sprintf("     [%s] chainId = %s", configFileYaml.Nodes[0].NodeUserIdent, configFileYaml.Genesis.Chain_Id))
}
