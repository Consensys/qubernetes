package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"strings"
)

// commands related to networking services.
var (
	urlGetCommand = cli.Command{
		Name:    "url",
		Usage:   "list url for node(s)/pod(s)",
		Aliases: []string{"urls"},
		Flags: []cli.Flag{
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
			// If no nodes were specified, look for all services containing "quorum".
			if len(nodeNames) == 0 {
				nodeNames = append(nodeNames, "quorum")
			}
			for _, nodeName := range nodeNames {
				serviceNames := serviceNamesFromPrefix(nodeName, namespace, false)
				for _, serviceName := range serviceNames {
					nodeServiceInfo := serviceInfoForNode(serviceName, urlType, namespace)
					if urlType == "nodeport" {
						if nodeServiceInfo.NodePortCakeshop != "" {
							fmt.Println("cakeshop - " + nodeIp + ":" + nodeServiceInfo.NodePortCakeshop)
						} else {
							fmt.Println(" geth      - " + nodeIp + ":" + nodeServiceInfo.NodePortGeth)
							fmt.Println(" tessera   - " + nodeIp + ":" + nodeServiceInfo.NodePortTm)
						}
					} else if urlType == "clusterip" { // the internal IP:Port of the specified node(s)
						if nodeServiceInfo.ClusterIPCakeshopURL != "" {
							//nodePort := nodePortForService(srvOut)
							fmt.Println("cakeshop - " + nodeServiceInfo.ClusterIPCakeshopURL)
						} else {
							fmt.Println(" geth      - " + nodeServiceInfo.ClusterIPGethURL)
							fmt.Println(" tessera   - " + nodeServiceInfo.ClusterIPTmURL)

						}
					}
				}
			}
			return nil
		},
	}
)

type NodeServiceInfo struct {
	ClusterIP string

	ClusterIPGethURL     string
	ClusterIPTmURL       string
	ClusterIPCakeshopURL string

	NodePortGeth     string
	NodePortTm       string
	NodePortCakeshop string
}

func serviceInfoForNode(nodeName, urlType, namespace string) NodeServiceInfo {
	//	fmt.Println("nodeName " + nodeName)
	var nodeServiceInfo NodeServiceInfo
	serviceNames := serviceNamesFromPrefix(nodeName, namespace, false)
	for _, serviceName := range serviceNames {
		serviceName = strings.TrimSpace(serviceName)
		srvOut := serviceForPrefix(serviceName, namespace, false)
		// NodePort will display the geth and tessera node ports for the specified node(s)
		// the nodePort can be accessed via the %Node_IP%:NodePort, the $NodeIP must be obtained
		// by the user, or outside this cli as various K8s have different ways of obtaining the $NodeIP, e.g.
		// minikube --> minikube ip
		// > qctl get url --type=nodeport | sed "s/<K8s_NODE_IP>/$(minikube ip)/g"
		// > qctl get url --type=nodeport --nodeip=$(minikube ip)
		if urlType == ServiceTypeNodePort {
			nodePortGeth := nodePortFormClusterPort(srvOut, DefaultGethPort)
			nodePortTessera := nodePortFormClusterPort(srvOut, DefaultTesseraPort)
			nodeServiceInfo.NodePortGeth = nodePortGeth
			nodeServiceInfo.NodePortTm = nodePortTessera
			if strings.Contains(serviceName, "cakeshop") {
				nodePort := nodePortForService(srvOut)
				nodeServiceInfo.NodePortCakeshop = nodePort
			}
		} else if urlType == ServiceTypeClusterIP { // the internal IP:Port of the specified node(s)
			clusterIp := clusterIpForService(srvOut)
			nodeServiceInfo.ClusterIP = clusterIp
			if strings.Contains(serviceName, "cakeshop") {
				nodePort := nodePortForService(srvOut)
				fmt.Println(serviceName + "- " + clusterIp + ":" + nodePort)
				nodeServiceInfo.ClusterIPCakeshopURL = clusterIp + ":" + nodePort
			} else {
				nodeServiceInfo.ClusterIPGethURL = clusterIp + ":" + DefaultGethPort
				nodeServiceInfo.ClusterIPTmURL = clusterIp + ":" + DefaultTesseraPort
				//fmt.Println(serviceName + " geth      - " + clusterIp + ":" + DefaultGethPort)
				//fmt.Println(serviceName + " tessera   - " + clusterIp + ":" + DefaultTesseraPort)
			}
		}
	}
	return nodeServiceInfo
}
