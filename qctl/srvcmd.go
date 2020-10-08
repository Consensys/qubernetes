package main

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

// commands related to networking services.
var (
	//  qctl ls url --node=cakeshop --node=quorum --node-ip=$(minikube ip)
	urlGetCommand = cli.Command{
		Name:    "url",
		Usage:   "list url for node(s)/pod(s)",
		Aliases: []string{"urls"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config, c",
				Usage:   "Load configuration from `FULL_PATH_FILE`",
				EnvVars: []string{"QUBE_CONFIG"},
				//Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "node, n",
				Usage: "node prefixes to retrieve service information from.`",
			},
			&cli.StringFlag{
				Name:  "type, t",
				Usage: "the type of URL to return, e.g. nodePort, clusterIp, ",
				Value: "clusterip",
			},
			&cli.StringFlag{
				Name:  "node-ip",
				Usage: "the IP of the K8s node, e.g. minikube ip ",
				Value: "<K8s_NODE_IP>",
			},
		},
		Action: func(c *cli.Context) error {
			namespace := c.String("namespace")
			nodeNames := c.StringSlice("node")
			nodeIp := c.String("node-ip")
			urlType := c.String("type")
			urlType = strings.ToLower(urlType)

			configFile := c.String("config")
			configFileYaml, err := LoadYamlConfig(configFile)
			if err != nil {
				log.Fatal("config file [%v] could not be loaded into the valid quebernetes yaml. err: [%v]", configFile, err)
			}

			// if no --node flags were set, display all quorum services known from the config.
			// if --node filter flags were set, only display the nodes that where set as a --node flag and
			// only if they exist in the config.
			allQuorumNodeK8sServices := mapset.NewSet()
			allQuorumOtherK8sServices := mapset.NewSet()
			nodeNamesFlags := mapset.NewSet()
			for _, nodeFlag := range nodeNames {
				nodeNamesFlags.Add(nodeFlag)
			}
			for i := 0; i < len(configFileYaml.Nodes); i++ {
				if nodeNamesFlags.Contains(configFileYaml.Nodes[i].NodeUserIdent) || len(nodeNames) == 0 {
					allQuorumNodeK8sServices.Add(configFileYaml.Nodes[i].NodeUserIdent)
				}
			}
			if configFileYaml.Cakeshop.Version != "" {
				if nodeNamesFlags.Contains("cakeshop") || len(nodeNames) == 0 {
					allQuorumOtherK8sServices.Add("cakeshop")
				}
			}
			if configFileYaml.Prometheus.Enabled == true {
				if nodeNamesFlags.Contains("monitor") || len(nodeNames) == 0 {
					allQuorumOtherK8sServices.Add("monitor")
				}
			}
			fmt.Println()
			// TODO: optimize this so we get all the services with one kubectl call then filter through the results.
			// display other quorum service first.
			for _, service := range allQuorumOtherK8sServices.ToSlice() {
				// need a string because ToSlice returns an []interface
				serviceName := fmt.Sprintf("%v", service)
				nodeServiceInfo := serviceInfoByPrefix(serviceName, urlType, namespace)
				if strings.Contains(serviceName, "monitor") { // monitor only support nodeport
					fmt.Println("prometheus server - " + nodeIp + ":" + nodeServiceInfo.NodePortPrometheus)
				} else if strings.Contains(serviceName, "cakeshop") { // cakeshop only support nodeport
					fmt.Println("cakeshop server - " + nodeIp + ":" + nodeServiceInfo.NodePortCakeshop)
				}
			}
			fmt.Println()

			// display all quorum node services.
			for _, service := range allQuorumNodeK8sServices.ToSlice() {
				// need a string because ToSlice returns an []interface
				serviceName := fmt.Sprintf("%v", service)
				nodeServiceInfo := serviceInfoByPrefix(serviceName, urlType, namespace)
				if urlType == "nodeport" {
					fmt.Println(serviceName + " geth      - " + nodeIp + ":" + nodeServiceInfo.NodePortGeth)
					fmt.Println(serviceName + " tessera   - " + nodeIp + ":" + nodeServiceInfo.NodePortTm)
				} else if urlType == "clusterip" { // the internal IP:Port of the specified node(s)
					fmt.Println(serviceName + " geth      - " + nodeServiceInfo.ClusterIPGethURL)
					fmt.Println(serviceName + " tessera   - " + nodeServiceInfo.ClusterIPTmURL)
				}
			}

			return nil
		},
	}
)

type NodeServiceInfo struct {
	ClusterIP string

	ClusterIPGethURL string
	ClusterIPTmURL   string
	//ClusterIPCakeshopURL string

	NodePortGeth       string
	NodePortTm         string
	NodePortCakeshop   string
	NodePortPrometheus string
}

func serviceInfoByPrefix(prefix, urlType, namespace string) NodeServiceInfo {
	//	fmt.Println("nodeName " + nodeName)
	var nodeServiceInfo NodeServiceInfo
	serviceNames := serviceNamesFromPrefix(prefix, namespace, false)
	for _, serviceName := range serviceNames {
		serviceName = strings.TrimSpace(serviceName)
		srvOut := serviceForPrefix(serviceName, namespace, false)
		if strings.Contains(serviceName, "monitor") { // only support nodeport
			nodePortProm := nodePortFormClusterPort(srvOut, DefaultPrometheusClusterPort)
			nodeServiceInfo.NodePortPrometheus = nodePortProm
		} else if strings.Contains(serviceName, "cakeshop") { // only support nodePort for now
			nodePort := nodePortForService(srvOut)
			nodeServiceInfo.NodePortCakeshop = nodePort
		} else {
			// NodePort will display the geth and tessera node ports for the specified node(s)
			// the nodePort can be accessed via the %Node_IP%:NodePort, the $NodeIP must be obtained
			// by the user, or outside this cli as various K8s have different ways of obtaining the $NodeIP, e.g.
			// minikube --> minikube ip
			// > qctl get url --type=nodeport | sed "s/<K8s_NODE_IP>/$(minikube ip)/g"
			// > qctl get url --type=nodeport --nodeip=$(minikube ip)
			if strings.ToLower(urlType) == strings.ToLower(ServiceTypeNodePort) {
				nodePortGeth := nodePortFormClusterPort(srvOut, DefaultGethPort)
				nodePortTessera := nodePortFormClusterPort(srvOut, DefaultTesseraPort)
				nodeServiceInfo.NodePortGeth = nodePortGeth
				nodeServiceInfo.NodePortTm = nodePortTessera
			} else if strings.ToLower(urlType) == strings.ToLower(ServiceTypeClusterIP) { // the internal IP:Port of the specified node(s)
				clusterIp := clusterIpForService(srvOut)
				nodeServiceInfo.ClusterIP = clusterIp
				nodeServiceInfo.ClusterIPGethURL = clusterIp + ":" + DefaultGethPort
				nodeServiceInfo.ClusterIPTmURL = clusterIp + ":" + DefaultTesseraPort
				//fmt.Println(serviceName + " geth      - " + clusterIp + ":" + DefaultGethPort)
				//fmt.Println(serviceName + " tessera   - " + clusterIp + ":" + DefaultTesseraPort)
			}
		}
	}
	return nodeServiceInfo
}
