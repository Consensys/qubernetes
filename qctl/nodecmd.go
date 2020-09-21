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
			if c.Args().Len() < 1 {
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

	nodeDeleteCommand = cli.Command{
		Name:  "node",
		Usage: "delete node and its associated resources (hard delete).",
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
			//&cli.StringFlag{
			//	Name:     "name",
			//	Usage:    "Unique name of node to delete",
			//	Required: true,
			//},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 1 {
				c.App.Run([]string{"qctl", "help", "delete", "node"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()
			fmt.Println("Delete node " + nodeName)
			// TODO: abstract this away as it is used in multiple places now.
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
				//displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull)
				if configFileYaml.Nodes[i].NodeUserIdent == nodeName {
					fmt.Println("Deleting node " + nodeName)
					// Remove K8s key resources if k8s dir set
					if k8sdir != "" {
						keyDirToDelete := configFileYaml.Nodes[i].KeyDir
						rmContents := exec.Command("rm", k8sdir+"/"+keyDirToDelete+"/*")
						fmt.Println(rmContents)
						rmdir := exec.Command("rmdir", k8sdir+"/"+keyDirToDelete)
						fmt.Println(rmdir)
						// TODO: delete  deployment, e.g. name: node5-deployment
					}
					configFileYaml.Nodes = append(configFileYaml.Nodes[:i], configFileYaml.Nodes[i+1:]...)
				}
			}

			// write file back
			WriteYamlConfig(configFileYaml, configFile)

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
				Usage:   "Transaction Manager Version.",
			},
			&cli.StringFlag{
				Name:  "tm",
				Usage: "Transaction Manager to user: tessera | constellation.",
			},
			&cli.StringFlag{
				Name:  "qimagefull",
				Usage: "The full repo + image name of the quorum image.",
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
			quorumImageFull := c.String("qimagefull")

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
			// check if the name is already taken
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				nodeEntry := configFileYaml.Nodes[i]
				if name == nodeEntry.NodeUserIdent {
					red.Println(fmt.Sprintf("Node name [%s] already exist!", name))
					displayNode("", nodeEntry, true, true, true, true, true, true, false, true)
					red.Println(fmt.Sprintf("Node name [%s] exists", name))
					return cli.Exit(fmt.Sprintf("Node name [%s] exists, node names must be unique", name), 3)
				}
			}
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
			nodeEntry := createNodeEntry(name, keyDir, consensus, quorumVersion, txManger, tmVersion, quorumImageFull)
			configFileYaml.Nodes = append(configFileYaml.Nodes, nodeEntry)
			fmt.Println()
			green.Println("Adding Node: ")
			displayNode("", nodeEntry, true, true, true, true, true, true, false, true)
			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			fmt.Println("The node(s) have been added to the config file [%s]", configFile)
			fmt.Println("Next, generate (update) the additional node resources for quorum and k8s:")
			fmt.Println()
			fmt.Println("**********************************************************************************************")
			fmt.Println()
			green.Println(fmt.Sprintf("  $> qctl generate network --update"))
			fmt.Println()
			fmt.Println("**********************************************************************************************")

			return nil
		},
	}
	// TODO: consolidate this and add node
	nodeUpdateCommand = cli.Command{
		Name:    "node",
		Usage:   "update node",
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
				Usage:   "Transaction Manager Version.",
			},
			&cli.StringFlag{
				Name:  "tm",
				Usage: "Transaction Manager to user: tessera | constellation.",
			},
			&cli.StringFlag{
				Name:  "qimagefull",
				Usage: "The full repo + image name of the quorum image.",
			},
			&cli.StringFlag{
				Name:  "gethparams",
				Usage: "Set the geth startup params.",
			},
		},
		Action: func(c *cli.Context) error {
			// defaults should be obtained from the config
			name := c.String("name")
			keyDir := c.String("keydir")
			if keyDir == "" {
				keyDir = fmt.Sprintf("key-%s", name)
			}
			//consensus := c.String("consensus")
			//quorumVersion := c.String("qversion")
			//tmVersion := c.String("tmversion")
			//txManger := c.String("tm")
			quorumImageFull := c.String("qimagefull")
			gethparams := c.String("gethparams")
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
			// find the nodes
			var updatedNode NodeEntry
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				nodeEntry := configFileYaml.Nodes[i]
				if name == nodeEntry.NodeUserIdent {
					displayNode("", nodeEntry, true, true, true, true, true, true, false, true)
					if gethparams != "" {
						nodeEntry.GethEntry.GetStartupParams = gethparams
					}
					if quorumImageFull != "" {
						nodeEntry.QuorumEntry.Quorum.DockerRepoFull = quorumImageFull
					}
					updatedNode = nodeEntry
					configFileYaml.Nodes[i] = updatedNode
					red.Println(fmt.Sprintf("updated nodes is [%v]", updatedNode))
				}
			}
			// If the node name the user entered to update does not exists, error out and notify the user.
			if updatedNode.NodeUserIdent == "" {
				red.Println(fmt.Sprintf("Node name [%s] does not exist, not updating any nodes.", name))
				fmt.Println()
				red.Println("to see current nodes run: ")
				fmt.Println()
				red.Println("  qctl ls nodes")
				fmt.Println()
				return cli.Exit(fmt.Sprintf("node name doesn't exist [%s]", name), 3)
			}
			fmt.Println(fmt.Sprintf("Updating node [%s] key dir [%s]", name, keyDir))
			fmt.Println()
			green.Println("Updating Node: ")
			displayNode("", updatedNode, true, true, true, true, true, true, false, true)
			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			fmt.Println("The node have been updated the config file [%s]", configFile)
			fmt.Println("Next, generate (update) the additional node resources for quorum and k8s:")
			fmt.Println()
			fmt.Println("**********************************************************************************************")
			fmt.Println()
			green.Println(fmt.Sprintf("  $> qctl generate network --update"))
			fmt.Println()
			fmt.Println("**********************************************************************************************")

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
			&cli.StringFlag{ // this is only required to get the enodeurl
				Name:    "k8sdir",
				Usage:   "The k8sdir (usually out) containing the output k8s resources",
				EnvVars: []string{"QUBE_K8S_DIR"},
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "display all node values",
			},
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
			isQuorumImageFull := c.Bool("qimagefull")
			isAll := c.Bool("all")
			k8sdir := c.String("k8sdir")
			// set all values to true
			if isAll {
				isName = true
				isConsensus = true
				isQuorumVersion = true
				isTmName = true
				isTmVersion = true
				if k8sdir != "" {
					isEnodeUrl = true
				}
				isQuorumImageFull = true
			}
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
			if k8sdir != "" {
				green.Println("  K8sdir set to:")
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
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			currentNum := len(configFileYaml.Nodes)
			fmt.Printf("config currently has %d nodes \n", currentNum)
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull)
			}

			return nil
		},
	}
)

func createNodeEntry(nodeName, nodeKeyDir, consensus, quorumVersion, txManger, tmVersion, quorumImageFull string) NodeEntry {
	quorum := Quorum{
		Consensus:      consensus,
		QuorumVersion:  quorumVersion,
		DockerRepoFull: quorumImageFull,
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

func displayNode(k8sdir string, nodeEntry NodeEntry, name, consensus, keydir, quorumVersion, txManger, tmVersion, isEnodeUrl, isQuorumImageFull bool) {
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
	if isQuorumImageFull {
		green.Println(fmt.Sprintf("     [%s] quorumImage: [%s]", nodeEntry.NodeUserIdent, nodeEntry.QuorumEntry.Quorum.DockerRepoFull))
	}
	if isEnodeUrl {
		if k8sdir == "" {
			red.Println("Set --k8sdir flag or QUBE_K8S_DIR env in order to display enodeurl")
		} else {
			enodeUrl := getEnodeId(nodeEntry.NodeUserIdent, k8sdir)
			if enodeUrl != "" {
				green.Println(fmt.Sprintf("     [%s] enodeUrl: [%s]", nodeEntry.NodeUserIdent, enodeUrl))
			}
		}
	}
	green.Println(fmt.Sprintf("     [%s] geth params: [%s]", nodeEntry.NodeUserIdent, nodeEntry.GethEntry.GetStartupParams))
	fmt.Println()
}
