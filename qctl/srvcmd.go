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
			},
		},
		Action: func(c *cli.Context) error {
			// TOOD: get these from the config file.
			gethPort := "8545"
			tesseraPort := "9080"
			namespace := c.String("namespace")
			nodeNames := c.StringSlice("node")
			nodeIp := c.String("node-ip")
			nodeIp = strings.ToLower(nodeIp)
			urlType := c.String("type")
			urlType = strings.ToLower(urlType)
			// If no nodes were specified, look for all services containing "quorum".
			if len(nodeNames) == 0 {
				nodeNames = append(nodeNames, "quorum")
			}
			for _, nodeName := range nodeNames {
				//	fmt.Println("nodeName " + nodeName)
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
					if urlType == "nodeport" {
						nodePortGeth := nodePortForClusterPort(srvOut, gethPort)
						nodePortTessera := nodePortForClusterPort(srvOut, tesseraPort)
						if strings.Contains(serviceName, "cakeshop") {
							nodePort := nodePortForService(srvOut)
							fmt.Println(serviceName + " - " + "<K8s_NODE_IP>" + ":" + nodePort)
						} else {
							if nodeIp != "" {
								fmt.Println(serviceName + " geth      - " + nodeIp + ":" + nodePortGeth)
								fmt.Println(serviceName + " tessera   - " + nodeIp + ":" + nodePortTessera)
							} else {
								fmt.Println(serviceName + " geth      - " + "<K8s_NODE_IP>" + ":" + nodePortGeth)
								fmt.Println(serviceName + " tessera   - " + "<K8s_NODE_IP>" + ":" + nodePortTessera)
							}
						}
					} else if urlType == "clusterip" { // the internal IP:Port of the specified node(s)
						//fmt.Println(fmt.Sprintf("srvOut: [%v]",srvOut))
						clusterIp := clusterIpForService(srvOut)
						//fmt.Println(fmt.Sprintf("Cluster IP: [%v]",clusterIp))
						if strings.Contains(serviceName, "cakeshop") {
							nodePort := nodePortForService(srvOut)
							fmt.Println(serviceName + "- " + clusterIp + ":" + nodePort)
						} else {
							fmt.Println(serviceName + " geth      - " + clusterIp + ":" + gethPort)
							fmt.Println(serviceName + " tessera   - " + clusterIp + ":" + tesseraPort)
						}
					}
				}
			}
			return nil
		},
	}
)
