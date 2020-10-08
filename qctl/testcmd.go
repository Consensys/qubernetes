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
	// qctl test contract --private node1
	testContractCmd = cli.Command{
		Name:    "contract",
		Aliases: []string{"contracts"},
		Usage:   "deploy the test contract(s) on a node.",
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

	// qctl test accepttest --node-ip=$(minikube ip)
	acceptanceWriteConfigCmd = cli.Command{
		Name:    "accepttest",
		Usage:   "output the acceptance test config file",
		Aliases: []string{"ac"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "k8sdir",
				Usage:    "k8sdir where the resources are stored, acceptance test config stored here for now.",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
			&cli.StringFlag{
				Name:  "node-ip",
				Usage: "the IP of the K8s node, e.g. minikube ip ",
				Value: "K8S_NODE_IP",
			},
		},
		Action: func(c *cli.Context) error {

			k8sNodeIp := c.String("node-ip")
			// TODO: abstract out config check used everywhere
			configFile := c.String("config")
			k8sdir := c.String("k8sdir")
			// get the current directory path, we'll use this in case the config file passed in was a relative path.
			pwdCmd := exec.Command("pwd")
			b := runCmd(pwdCmd)
			pwd := strings.TrimSpace(b.String())

			if configFile == "" {
				c.App.Run([]string{"qctl", "help", "accepttest"})

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
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}
			acceptanceTestYaml := createAcceptanceTestConfigString(configFileYaml, k8sNodeIp)
			fmt.Println(acceptanceTestYaml)
			//fmt.Println(config.ToString())
			configBytes := []byte(acceptanceTestYaml)
			acceptanceTestYamlFile := k8sdir + "/config/application-qctl-generated.yml"
			// try writing the generated file out to disk this file will be used to initialize the network.
			// TODO: it might be best to store in K8s itself
			err = ioutil.WriteFile(acceptanceTestYamlFile, configBytes, 0644)
			if err != nil {
				log.Fatal("error writing acceptanceTestYamlFil to [%v]. err: [%v]", acceptanceTestYamlFile, err)
			}
			return nil
		},
	}
	// qctl test accepttest --node-ip=$(minikube ip)
	// docker run --rm -v $(pwd):/tmp/config -e "SPRING_CONFIG_ADDITIONALLOCATION=file:/tmp/config/" -e "SPRING_PROFILES_ACTIVE=${PROFILE}" quorumengineering/acctests:latest test  -Dtags="basic && !externally-signed && !personal-api-signed && !eth-api-signed"
	acceptanceTestRunCmd = cli.Command{
		Name:    "accepttest",
		Aliases: []string{"accepttests, acceptancetest, acceptancetests"},
		Usage:   "run the acceptance tests.",
		// TODO: pass in tags, and the config file, will also need the k8s ip...unless
		// additionally pass in the file, and the tags, then no additional config needed
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config, c",
				Usage:    "Load configuration from `FULL_PATH_FILE`",
				EnvVars:  []string{"QUBE_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "k8sdir",
				Usage:    "k8sdir where the resources are stored, acceptance test config stored here for now.",
				EnvVars:  []string{"QUBE_K8S_DIR"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "node-ip",
				Usage:    "the IP of the K8s node, e.g. minikube ip ",
				Value:    "K8S_NODE_IP",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "tags",
				Aliases: []string{"t", "tag"},
				Usage:   "tags indicating which test to run, if not set, defaults to: (basic || basic-{CONSENSUS} || networks/typical::{CONSENSUS}) && !extension ",
			},
		},
		Action: func(c *cli.Context) error {

			k8sdir := c.String("k8sdir")
			k8sNodeIp := c.String("node-ip")

			if k8sdir == "" {
				c.App.Run([]string{"qctl", "help", "accepttest"})

				// QUBE_CONFIG or flag
				fmt.Println()

				fmt.Println()
				red.Println("  --k8sdir flag must be provided.")
				red.Println("             or ")
				red.Println("     QUBE_K8S_DIR environment variable needs to be set to your k8sdir.")
				fmt.Println()
			}

			// TODO: abstract out config check used everywhere
			configFile := c.String("config")
			// V1 keep the file on a known place on disc
			// acceptanceTestYaml := k8sdir + "/config/application-qctl-generated.yml"
			// end V1
			configFileYaml, err := LoadYamlConfig(configFile)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}

			// acceptance test file must be prefixed with application, e.g. `application-$MYNAME.yaml`

			// if an acceptance test config file wasn't provided, create one against the running network now.
			acceptanceTestYaml := createAcceptanceTestConfigString(configFileYaml, k8sNodeIp)
			configBytes := []byte(acceptanceTestYaml)
			// try writing the generated file out to disk this file will be used to initialize the network.
			// TODO: it might be best to store in K8s itself
			acceptanceTestYamlFile := k8sdir + "/config/application-qctl-generated.yml"
			err = ioutil.WriteFile(acceptanceTestYamlFile, configBytes, 0644)
			if err != nil {
				log.Fatal("error writing acceptanceTestYamlFil to [%v]. err: [%v]", acceptanceTestYamlFile, err)
			}

			acceptanceTestProfile := "qctl-generated"
			// Depending on the consensus (raft or istanbul) of the network, set the tags accordingly.
			tags := c.String("tags")
			if tags == "" {
				if configFileYaml.Genesis.Consensus == RaftConsensus {
					tags = "(basic || basic-raft || networks/typical::raft) && !extension"
				} else {
					tags = "(basic || basic-istanbul || networks/typical::istanbul) && !extension"
				}
			}
			green.Println("using profile file: " + acceptanceTestYamlFile)
			// if debugging include -X for mvn output
			cmd := exec.Command("docker", "run", "--rm", "-v", k8sdir+"/config:/tmp/config", "-e", "SPRING_CONFIG_ADDITIONALLOCATION=file:/tmp/config/", "-e", "SPRING_PROFILES_ACTIVE="+acceptanceTestProfile, "quorumengineering/acctests:latest", "test", "-Dtags="+tags)
			// e.g. docker run --rm -v /Users/libby/Workspace.Quorum/qctl-config/out/config:/tmp/config -e SPRING_CONFIG_ADDITIONALLOCATION=file:/tmp/config/ -e SPRING_PROFILES_ACTIVE=qctl-generated quorumengineering/acctests:latest test -Dtags='(basic || basic-istanbul || networks/typical::istanbul) && !extension'
			fmt.Println(cmd)
			dropIntoCmd(cmd)
			return nil
		},
	}
)

// acceptance test expects a config file with node names labelled in sequential order, Node1, Node2:
//quorum:
//nodes:
//Node1:
//	privacy-address: WC6yWjDXG9uFQTTfV+bkTr5GbqjmH7DotcOYqeSajgs=
//		url: http://192.168.64.49:32507
//	third-party-url: http://192.168.64.49:31375
//Node2:
//	privacy-address: t+lu+T9VoTWfQMyF7wiFsiEKARaQqgvOJynNfrsSbAk=
//		url: http://192.168.64.49:32106
//	third-party-url: http://192.168.64.49:30968
//
// TODO: can we update the format to use unique names, and yaml lists.
//quorum:
//  nodes:
//  - name: quorum-node1
//    privacy-address: WC6yWjDXG9uFQTTfV+bkTr5GbqjmH7DotcOYqeSajgs=
//    url: http://192.168.64.49:32507
//    third-party-url: http://192.168.64.49:31375
//  - name: quorum-node2
//    privacy-address: t+lu+T9VoTWfQMyF7wiFsiEKARaQqgvOJynNfrsSbAk=
//    url: http://192.168.64.49:32106
//    third-party-url: http://192.168.64.49:30968
func createAcceptanceTestConfigString(configFileYaml QConfig, k8sNodeIp string) string {
	acceptanceTestYaml := AcceptTestConfig{}
	for _, node := range configFileYaml.Nodes {
		// get nodes from config, and set the necessary tm public key, geth and tessera urls, using NodePort.
		serviceNodePort := serviceInfoByPrefix(node.NodeUserIdent, ServiceTypeNodePort, "")
		tmPublicKey := getTmPublicKey(node.NodeUserIdent)
		nodeEntry := ATNodeEntry{}
		nodeEntry.GethURL = "http://" + k8sNodeIp + ":" + serviceNodePort.NodePortGeth
		nodeEntry.TmPublicKey = tmPublicKey
		nodeEntry.TmURL = "http://" + k8sNodeIp + ":" + serviceNodePort.NodePortTm

		acceptanceTestYaml.Quorum.Nodes = append(acceptanceTestYaml.Quorum.Nodes, nodeEntry)
	}
	// TODO this is really hideous, better to change the acceptance test to take a yaml file that has a list of node with a name attributes:
	//quorum:
	//  nodes:
	//  - name: node1
	//    privacy-address: WC6yWjDXG9uFQTTfV+bkTr5GbqjmH7DotcOYqeSajgs=
	//    url: http://192.168.64.49:32507
	//    third-party-url: http://192.168.64.49:31375
	//  - name: node2
	//    privacy-address: t+lu+T9VoTWfQMyF7wiFsiEKARaQqgvOJynNfrsSbAk=
	//    url: http://192.168.64.49:32106
	//    third-party-url: http://192.168.64.49:30968
	acceptanceTestUpdatedYaml := acceptanceTestYaml.ToString()
	// since node name must be in the for form Node(i+1)
	for i := 0; i < len(configFileYaml.Nodes); i++ {
		//acceptanceTestUpdatedYaml = strings.Replace(acceptanceTestUpdatedYaml, "- privacy-address:", "  " + node.NodeUserIdent + ": \n      privacy-address:", 1)
		nodeName := fmt.Sprintf("Node%d", i+1)
		acceptanceTestUpdatedYaml = strings.Replace(acceptanceTestUpdatedYaml, "- privacy-address:", "  "+nodeName+": \n      privacy-address:", 1)
	}
	acceptanceTestUpdatedYaml = strings.ReplaceAll(acceptanceTestUpdatedYaml, " url:", "   url:")
	acceptanceTestUpdatedYaml = strings.ReplaceAll(acceptanceTestUpdatedYaml, "third-party-url:", "  third-party-url:")
	return acceptanceTestUpdatedYaml
}
