package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"os/exec"
	"strings"
)

var (
	nodeConnectCommand = cli.Command{
		Name:      "connect",
		Aliases:   []string{"c"},
		Usage:     "connect to nodes / pods",
		ArgsUsage: "[pod_substring] [quorum | tessera | constellation]",
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 2 {
				c.App.Run([]string{"qctl", "help", "connect", "node"})
				return cli.Exit("wrong number of arguments", 2)
			}
			namespace := c.String("namespace")
			nodeName := c.Args().First()
			container := c.Args().Get(1)
			if container == "" {
				container = "quorum"
			}
			podName := podNameFromPrefix(nodeName, namespace)
			log.Printf("trying to connect pods [%v]", podName)
			cmd := exec.Command("kubectl", "--namespace="+namespace, "exec", "-it", podName, "-c", container, "--", "/bin/ash")
			dropIntoCmd(cmd)
			return nil
		},
	}
	//qctl add node --id=node3 --consensus=ibft --quorum
	//TODO: get the defaults from the config file.
	nodeAddCommand = cli.Command{
		Name:    "node",
		Usage:   "add new nodes",
		Aliases: []string{"n", "nodes"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FULL_PATH_FILE`",
				EnvVars: []string{"QUBE_CONFIG"},
				//Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Unique name of node to create",
				Required: true,
			},
			// TODO: set default to Node-name-key-dir
			&cli.StringFlag{
				Name:  "keydir",
				Usage: "key dir where the newly generated key will be placed",
			},
			&cli.StringFlag{
				Name:  "consensus",
				Usage: "Consensus to use raft | istanbul.",
			},
			&cli.StringFlag{
				Name:    "qversion",
				Aliases: []string{"qv"},
				Usage:   "Quorum Version.",
			},
			&cli.StringFlag{
				Name:    "tmversion",
				Aliases: []string{"tmv"},
				Usage:   "Transaction Manger Version.",
			},
			&cli.StringFlag{
				Name:  "tm",
				Usage: "Transaction Manger to user: tessera | constellation.",
			},
		},
		Action: func(c *cli.Context) error {
			// defaults should be obtained from the config
			name := c.String("name")
			keyDir := c.String("keydir")
			if keyDir == "" {
				keyDir = fmt.Sprintf("key-%s", name)
			}
			consensus := c.String("consensus")
			quorumVersion := c.String("qversion")
			tmVersion := c.String("tmversion")
			txManger := c.String("tm")

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
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			// set defaults from the existing config if node values were not provided
			if quorumVersion == "" {
				quorumVersion = configFileYaml.Genesis.QuorumVersion
			}
			if consensus == "" {
				consensus = configFileYaml.Genesis.Consensus
			}
			// for the transaction manager, set the defaults to what is available on the first node.
			if txManger == "" {
				txManger = configFileYaml.Nodes[0].QuorumEntry.Tm.Name
			}
			if tmVersion == "" {
				tmVersion = configFileYaml.Nodes[0].QuorumEntry.Tm.TmVersion
			}
			fmt.Println(fmt.Sprintf("Adding node [%s] key dir [%s]", name, keyDir))
			currentNum := len(configFileYaml.Nodes)
			fmt.Println(fmt.Sprintf("config currently has %d nodes", currentNum))
			nodeEntry := createNodeEntry(name, keyDir, consensus, quorumVersion, txManger, tmVersion)
			configFileYaml.Nodes = append(configFileYaml.Nodes, nodeEntry)
			fmt.Println()
			green.Println("Adding Node: ")
			displayNode("", nodeEntry, true, true, true, true, true, true, false)
			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			return nil
		},
	}
	// qctl ls node --name --consensus --quorumversion
	// qctl ls node --name --consensus --quorumversion --tmversion --tmname
	nodeListCommand = cli.Command{
		Name:    "node",
		Usage:   "list nodes info",
		Aliases: []string{"n", "nodes"},
		Flags: []cli.Flag{

			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{ // this is only required to get the enodeId
				Name:     "k8sdir",
				Usage:    "The k8sdir (usually out) containing the output k8s resources",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "display all node values",
			},
			// TODO: have a --filter=name,consensus,quorumversion
			&cli.BoolFlag{
				Name:  "name",
				Usage: "display the name of the node",
			},
			&cli.BoolFlag{
				Name:  "consensus",
				Usage: "display the consensus of the node",
			},
			&cli.BoolFlag{
				Name:  "quorumversion",
				Usage: "display the quorumversion of the node",
			},
			&cli.BoolFlag{
				Name:  "tmname",
				Usage: "display the tm name of the node",
			},
			&cli.BoolFlag{
				Name:  "tmversion",
				Usage: "display the tm version of the node",
			},
			&cli.BoolFlag{
				Name:  "keydir",
				Usage: "display the keydir of the node",
			},
			&cli.BoolFlag{
				Name:  "enodeurl",
				Usage: "display the enodeurl of the node",
			},
		},
		Action: func(c *cli.Context) error {

			isName := c.Bool("name")
			isConsensus := c.Bool("consensus")
			isQuorumVersion := c.Bool("quorumversion")
			isTmName := c.Bool("tmname")
			isTmVersion := c.Bool("tmversion")
			isKeyDir := c.Bool("keydir")
			isEnodeUrl := c.Bool("enodeurl")
			isAll := c.Bool("all")
			// set all values to true
			if isAll {
				isName = true
				isConsensus = true
				isQuorumVersion = true
				isTmName = true
				isTmVersion = true
			}
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
			configFileYaml, err := LoadYamlConfig(configFile)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			currentNum := len(configFileYaml.Nodes)
			fmt.Printf("config currently has %d nodes \n", currentNum)
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl)
			}

			return nil
		},
	}
)

func createNodeEntry(nodeName, nodeKeyDir, consensus, quorumVersion, txManger, tmVersion string) NodeEntry {
	quorum := Quorum{
		Consensus:     consensus,
		QuorumVersion: quorumVersion,
	}
	tm := Tm{
		Name:      txManger,
		TmVersion: tmVersion,
	}
	quorumEntry := QuorumEntry{
		Quorum: quorum,
		Tm:     tm,
	}
	nodeEntry := NodeEntry{
		NodeUserIdent: nodeName,
		KeyDir:        nodeKeyDir,
		QuorumEntry:   quorumEntry,
	}
	return nodeEntry
}

// QUBE_K8S_DIR
// cat $QUBE_K8S_DIR/config/permissioned-nodes.json | grep quorum-node1
func getEnodeId(nodeName, qubeK8sDir string) string {
	c1 := exec.Command("cat", qubeK8sDir+"/config/permissioned-nodes.json")
	c2 := exec.Command("grep", nodeName)

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var out bytes.Buffer
	c2.Stdout = &out
	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()
	enodeUrl := strings.TrimSpace(out.String())
	return enodeUrl
}

func displayNode(k8sdir string, nodeEntry NodeEntry, name, consensus, keydir, quorumVersion, txManger, tmVersion, isEnodeUrl bool) {
	fmt.Println()
	green.Println(fmt.Sprintf("     [%s]", nodeEntry.NodeUserIdent))
	if keydir {
		green.Println(fmt.Sprintf("     [%s] keydir: [%s]", nodeEntry.NodeUserIdent, nodeEntry.KeyDir))
	}
	if consensus {
		green.Println(fmt.Sprintf("     [%s] consensus: [%s]", nodeEntry.NodeUserIdent, nodeEntry.QuorumEntry.Quorum.Consensus))
	}
	if quorumVersion {
		green.Println(fmt.Sprintf("     [%s] quorumVersion: [%s]", nodeEntry.NodeUserIdent, nodeEntry.QuorumEntry.Quorum.QuorumVersion))
	}
	if txManger {
		green.Println(fmt.Sprintf("     [%s] txManger: [%s]", nodeEntry.NodeUserIdent, nodeEntry.QuorumEntry.Tm.Name))
	}
	if tmVersion {
		green.Println(fmt.Sprintf("     [%s] tmVersion: [%s]", nodeEntry.NodeUserIdent, nodeEntry.QuorumEntry.Tm.TmVersion))
	}
	if isEnodeUrl {
		enodeUrl := getEnodeId(nodeEntry.NodeUserIdent, k8sdir)
		if enodeUrl != "" {
			green.Println(fmt.Sprintf("     [%s] enodeUrl: [%s]", nodeEntry.NodeUserIdent, enodeUrl))
		}
	}
	fmt.Println()
}
