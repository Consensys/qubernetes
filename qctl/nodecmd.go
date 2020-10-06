package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"os"
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
	// qctl delete node --hard  quorum-node5
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
			&cli.BoolFlag{ // this is only required if the user wants to delete the generated (k8s/quorum) resources as well.
				Name:  "hard",
				Usage: "delete all associated resources with this node, e.g. keys, configs, etc.",
				Value: false,
			},
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
			isHardDelete := c.Bool("hard")
			// get the current directory path, we'll use this in case the config file passed in was a relative path.
			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

			if configFile == "" {
				c.App.Run([]string{"qctl", "help", "init"})
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
			currentNum := len(configFileYaml.Nodes)
			fmt.Printf("config currently has %d nodes \n", currentNum)
			var nodeToDelete NodeEntry
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				//displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull)
				if configFileYaml.Nodes[i].NodeUserIdent == nodeName {
					fmt.Println("Deleting node " + nodeName)
					nodeToDelete = configFileYaml.Nodes[i]
					// try to remove the running k8s resources
					stopNode(nodeName)
					rmPersistentData(nodeName)
					rmService(nodeName)
					// TEST, if it is raft, remove it from the cluster
					if configFileYaml.Nodes[i].QuorumEntry.Quorum.Consensus == "raft" {
						// TODO: find a running node? it could either be the previous node or next node, check the index.
						// run raft.removePeer(raftId)
					}
					// Delete the resources files associated with the node, e.g. keys, k8s files, etc.
					if k8sdir != "" && isHardDelete {
						red.Println("Is hard delete remove key files and directory")
						keyDirToDelete := configFileYaml.Nodes[i].KeyDir
						nodeToDeleteKeyDir := k8sdir + "/config/" + keyDirToDelete

						// TODO: hard delete delete keys
						//rmContents := exec.Command("rm", "-f", nodeToDeleteKeyDir+"/*")
						// explicitly delete all the files that should be in the directory.
						rmContents := exec.Command("rm", "-f", nodeToDeleteKeyDir+"/acctkeyfile.json")
						dropIntoCmd(rmContents)
						rmContents = exec.Command("rm", "-f", nodeToDeleteKeyDir+"/enode")
						dropIntoCmd(rmContents)
						rmContents = exec.Command("rm", "-f", nodeToDeleteKeyDir+"/nodekey")
						dropIntoCmd(rmContents)
						rmContents = exec.Command("rm", "-f", nodeToDeleteKeyDir+"/password.txt")
						dropIntoCmd(rmContents)
						rmContents = exec.Command("rm", "-f", nodeToDeleteKeyDir+"/tm.key")
						dropIntoCmd(rmContents)
						rmContents = exec.Command("rm", "-f", nodeToDeleteKeyDir+"/tm.pub")
						dropIntoCmd(rmContents)
						// instead of running  rm -r, run rmdir on what should be an empty dir,
						// rmdir will return an error if the directory doesn't exist, so check if dir exists first.
						_, err := os.Stat(nodeToDeleteKeyDir)
						if os.IsNotExist(err) {
							log.Fatal(fmt.Sprintf("Directory does not exist, ignoring dir [%s]", nodeToDeleteKeyDir))
						} else {
							rmdir := exec.Command("rmdir", nodeToDeleteKeyDir)
							fmt.Println(rmdir)
							dropIntoCmd(rmdir)
						}

						//rmdir := exec.Command("rm", "-r", "-f", nodeToDeleteKeyDir)

					}
					// TODO: delete k8s deployment file, e.g. name: quorum-node5-quorum-deployment.yaml
					rmDeploymentFile := exec.Command("rm", "-f", k8sdir+"/deployments/"+nodeToDelete.NodeUserIdent+"-quorum-deployment.yaml")
					runCmd(rmDeploymentFile)
					// finally remove the node from the the qubernetes config, if the resources have not been delete,
					// it can be added back using the old name and it will use the keys that have not been deleted.
					configFileYaml.Nodes = append(configFileYaml.Nodes[:i], configFileYaml.Nodes[i+1:]...)
				}
			}

			// write file back
			WriteYamlConfig(configFileYaml, configFile)
			green.Println(fmt.Sprintf("  Deleted node [%s]", nodeToDelete.NodeUserIdent))
			if nodeToDelete.QuorumEntry.Quorum.Consensus == "raft" {
				green.Println(fmt.Sprintf("  This was raft node, and has not been removed from the cluster. "))
				green.Println(fmt.Sprintf("  To remove it from the current raft cluster, run on an healthy node: "))
				green.Println(fmt.Sprintf("  qctl geth exec node1 'raft.cluster'"))
				green.Println(fmt.Sprintf("  qctl geth exec node1 'raft.removePeer()'"))
			}

			return nil
		},
	}

	/*
	 * stops the give node, stopping will only remove the deployment from the K8s cluster, it will not remove any other
	 * associated resources, such as the PVC (persistent volume claim) therefore maintaining the state of the node. The
	 * services, key, and other resources are kept.
	 * The node can be restarted again, by running `qctl network create`
	 *
	 * > qctl stop node quorum-node5
	 */
	nodeStopCommand = cli.Command{
		Name:  "node",
		Usage: "stop the node(s) by deleting the associated K8s deployment.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() < 1 {
				c.App.Run([]string{"qctl", "help", "stop", "node"})
				return cli.Exit("wrong number of arguments", 2)
			}
			nodeName := c.Args().First()

			// TODO: abstract this away as it is used in multiple places now.
			configFile := c.String("config")
			// get the current directory path, we'll use this in case the config file passed in was a relative path.
			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

			if configFile == "" {
				c.App.Run([]string{"qctl", "help", "init"})
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
			currentNum := len(configFileYaml.Nodes)
			fmt.Printf("config currently has %d nodes \n", currentNum)
			var nodeToStop NodeEntry
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				if configFileYaml.Nodes[i].NodeUserIdent == nodeName {
					fmt.Println("Stopping node " + nodeName)
					nodeToStop = configFileYaml.Nodes[i]
					// try to remove the running k8s resources
					stopNode(nodeName)
					green.Println(fmt.Sprintf("  Stopped node [%s]", nodeToStop.NodeUserIdent))
					green.Println()
					green.Println("  to restart node run: ")
					green.Println()
					green.Println(fmt.Sprintf("    qctl deploy network"))
					green.Println()
					if nodeToStop.QuorumEntry.Quorum.Consensus == "raft" {
						green.Println(fmt.Sprintf("  This was raft node, and has not been removed from the cluster. "))
						green.Println(fmt.Sprintf("  To remove it from the current raft cluster, run on an healthy node: "))
						green.Println(fmt.Sprintf("  qctl geth exec node1 'raft.cluster'"))
						green.Println(fmt.Sprintf("  qctl geth exec node1 'raft.removePeer()'"))
					}
				}
			}
			if nodeToStop.NodeUserIdent == "" {
				fmt.Println()
				red.Println(fmt.Sprintf("  Node [%s] not found in config", nodeName))
				fmt.Println()
				red.Println(fmt.Sprintf("  To list nodes run:"))
				fmt.Println()
				red.Println("    qctl ls nodes ")
				fmt.Println()
			}

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
					displayNode("", nodeEntry, true, true, true, true, true, true, false, true, true)
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
			displayNode("", nodeEntry, true, true, true, true, true, true, false, true, true)
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
					displayNode("", nodeEntry, true, true, true, true, true, true, false, true, true)
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
			displayNode("", updatedNode, true, true, true, true, true, true, false, true, true)
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
				Name:    "enodeurl",
				Aliases: []string{"enode"},
				Usage:   "display the enodeurl of the node",
			},
			&cli.BoolFlag{
				Name:    "gethparams",
				Aliases: []string{"gp"},
				Usage:   "display the geth startup params of the node",
			},
			&cli.BoolFlag{
				Name:    "bare",
				Aliases: []string{"b"},
				Usage:   "display the minimum output, useful for scripts / automation",
			},
		},
		Action: func(c *cli.Context) error {
			// potentially show only this node
			nodeName := c.Args().First()
			nodeFound := true
			if nodeName != "" { // if the user request a specific node, we want to make sure we let them know it was found or not.
				nodeFound = false
			}
			isName := c.Bool("name")
			isConsensus := c.Bool("consensus")
			isQuorumVersion := c.Bool("quorumversion")
			isTmName := c.Bool("tmname")
			isTmVersion := c.Bool("tmversion")
			isKeyDir := c.Bool("keydir")
			isEnodeUrl := c.Bool("enodeurl")
			isQuorumImageFull := c.Bool("qimagefull")
			isGethParams := c.Bool("gethparams")
			isAll := c.Bool("all")
			isBare := c.Bool("bare")
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
				isGethParams = true
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
			if !isBare {
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
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			currentNum := len(configFileYaml.Nodes)
			if !isBare {
				fmt.Printf("config currently has %d nodes \n", currentNum)
			}

			for i := 0; i < len(configFileYaml.Nodes); i++ {
				if nodeName == "" { // node name not set always show node
					if isBare { // show the bare version, cleaner for scripts.
						displayNodeBare(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull, isGethParams)
					} else {
						displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull, isGethParams)
					}
				} else if nodeName == configFileYaml.Nodes[i].NodeUserIdent {
					nodeFound = true
					if isBare { // show the bare version, cleaner for scripts.
						displayNodeBare(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull, isGethParams)
					} else {
						displayNode(k8sdir, configFileYaml.Nodes[i], isName, isKeyDir, isConsensus, isQuorumVersion, isTmName, isTmVersion, isEnodeUrl, isQuorumImageFull, isGethParams)
					}
				}
			}
			// if the nodename was specified, but not found in the config, list the names of the nodes for the user.
			if !nodeFound {
				fmt.Println()
				red.Println(fmt.Sprintf("  Node name [%s] not found in config file ", nodeName))
				fmt.Println()
				fmt.Println(fmt.Sprintf("  Node Names are: "))
				for i := 0; i < len(configFileYaml.Nodes); i++ {
					fmt.Println(fmt.Sprintf("    [%s]", configFileYaml.Nodes[i].NodeUserIdent))
				}
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
	enodeUrl = strings.ReplaceAll(enodeUrl, ",", "")
	return enodeUrl
}

func displayNode(k8sdir string, nodeEntry NodeEntry, name, consensus, keydir, quorumVersion, txManger, tmVersion, isEnodeUrl, isQuorumImageFull, isGethParms bool) {
	fmt.Println()
	green.Println(fmt.Sprintf("     [%s] unique name", nodeEntry.NodeUserIdent))
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
	if isGethParms {
		green.Println(fmt.Sprintf("     [%s] geth params: [%s]", nodeEntry.NodeUserIdent, nodeEntry.GethEntry.GetStartupParams))
	}
	fmt.Println()
}

func displayNodeBare(k8sdir string, nodeEntry NodeEntry, name, consensus, keydir, quorumVersion, txManger, tmVersion, isEnodeUrl, isQuorumImageFull, isGethParms bool) {
	if name {
		fmt.Println(nodeEntry.NodeUserIdent)
	}
	if keydir {
		fmt.Println(nodeEntry.KeyDir)
	}
	if consensus {
		fmt.Println(nodeEntry.QuorumEntry.Quorum.Consensus)
	}
	if quorumVersion {
		fmt.Println(nodeEntry.QuorumEntry.Quorum.QuorumVersion)
	}
	if txManger {
		fmt.Println(nodeEntry.QuorumEntry.Tm.Name)
	}
	if tmVersion {
		fmt.Println(nodeEntry.QuorumEntry.Tm.TmVersion)
	}
	if isQuorumImageFull {
		fmt.Println(nodeEntry.QuorumEntry.Quorum.DockerRepoFull)
	}
	if isEnodeUrl {
		if k8sdir == "" {
			red.Println("Set --k8sdir flag or QUBE_K8S_DIR env in order to display enodeurl")
		} else {
			enodeUrl := getEnodeId(nodeEntry.NodeUserIdent, k8sdir)
			fmt.Println(enodeUrl)
		}
	}
	if isGethParms {
		fmt.Println(nodeEntry.GethEntry.GetStartupParams)
	}
}

// stop node should just remove the deployment, and not delete any resources or persistent data.
func stopNode(nodeName string) error {
	// TODO: should there be a separate delete and remove node? where remove only removes it from the cluster, but delete removes all traces?
	rmRunningDeployment := exec.Command("kubectl", "delete", "deployment", nodeName+"-deployment")
	fmt.Println(rmRunningDeployment)
	// TODO: run should return the error so we can handle it or ignore it.
	var out bytes.Buffer
	rmRunningDeployment.Stdout = &out
	err := rmRunningDeployment.Run()
	if err != nil { // log the error but don't throw any
		log.Info("deployment not found in k8s, ignoring.")
	}
	return err
}

// TODO: handle errors, etc.
func rmPersistentData(nodeName string) error {
	// remove the persistent data.
	rmPVC := exec.Command("kubectl", "delete", "pvc", nodeName+"-pvc")
	fmt.Println(rmPVC)
	var out bytes.Buffer
	rmPVC.Stdout = &out
	err := rmPVC.Run()
	if err != nil { // log the error but don't throw any
		log.Info("PVC / Persistent data not found in k8s, ignoring.")
	}
	return err
}

func rmService(nodeName string) error {
	// remove the persistent data.
	rmService := exec.Command("kubectl", "delete", "service", nodeName)
	fmt.Println(rmService)
	var out bytes.Buffer
	rmService.Stdout = &out
	err := rmService.Run()
	if err != nil { // log the error but don't throw any
		log.Info("service not found in k8s, ignoring.")
	}
	return err
}

func getTmPublicKey(nodeName string) string {
	//c1 := exec.Command("cat", qubeK8sDir+"/config/" + nodeKeyDir + "tm.pub")
	//kc get configMaps quorum-node3-tm-key-config -o yaml | grep "tm.pub:"
	c1 := exec.Command("kubectl", "get", "configMap", nodeName+"-tm-key-config", "-o", "yaml")
	c2 := exec.Command("grep", "tm.pub:")

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
	// output will look like:
	// tm.pub: dF+Y81qRKI3Noh6ldI+FnQmqmjRYvOqLCaooTi5txi4=
	tmPublicKey := strings.ReplaceAll(out.String(), "tm.pub:", "")
	tmPublicKey = strings.TrimSpace(tmPublicKey)
	return tmPublicKey
}
