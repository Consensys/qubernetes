package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	red   = color.New(color.FgRed)
	green = color.New(color.FgGreen)

	DefaultQuorumVersion        = "2.6.0"
	DefaultTmName               = "tessera"
	DefaultTesseraVersion       = "0.10.4"
	DefaultConstellationVersion = "0.3.2"
	DefaultConesensus           = "istanbul"
	DefaultNodeNumber           = 4
	DefaultChainId              = "1000"
)

func podNameFromPrefix(prefix string, namespace string) string {
	//log.Printf("connecting to node [%v] ", prefix)
	//TODO: extract this into a utils function
	c1 := exec.Command("kubectl", "--namespace="+namespace, "get", "pods")
	fmt.Println(c1.String())

	c2 := exec.Command("grep", prefix)
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2
	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	//var out bytes.Buffer
	//cmd.Stdout = &out
	//err := cmd.Run()
	//if err != nil {
	//	log.Fatal(err)
	//}
	podOutput := b2.String()
	//io.Copy(os.Stdout, &b2)
	//fmt.Printf(podOutput)
	podName := strings.Split(podOutput, " ")[0]
	return podName
}

func serviceNamesFromPrefix(prefix string, namespace string, info bool) []string {
	//log.Printf("connecting to node [%v] ", prefix)
	//TODO: extract this into a utils function
	c1 := exec.Command("kubectl", "--namespace="+namespace, "get", "service")
	if info {
		fmt.Println(c1.String())
	}
	c2 := exec.Command("grep", prefix)
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2
	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	srvOutput := b2.String()
	//fmt.Println("srvOutput", srvOutput)
	// split output on new line, this will add an extra empty entry in the array, e.g. if 1 item is returned, there
	// will be 2 items in the array.
	srvNames := strings.Split(srvOutput, "\n")
	srvNames = srvNames[:len(srvNames)-1]
	//fmt.Println("srvNames", srvNames)
	//fmt.Println("There are ", len(srvNames))
	//serviceCt := len(srvNames)
	var names []string
	for _, s := range srvNames {
		name := strings.Split(s, " ")[0]
		//fmt.Println("name: ", name)
		names = append(names, name)
	}
	return names
	// if the result returns more than one service entry this will only return the first one.
	//srvName := strings.Split(srvOutput, " ")[0]
	//return srvName
}

func serviceForPrefix(prefix string, namespace string, info bool) string {
	c1 := exec.Command("kubectl", "--namespace="+namespace, "get", "service")
	if info {
		fmt.Println(c1.String())
	}
	c2 := exec.Command("grep", prefix)
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2
	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	srvOutput := b2.String()
	return srvOutput
}

// get the clusterIP for the given service
// serviceOutput is the output of kubectl get service for a single service.
func clusterIpForService(serviceOutputStr string) string {
	c1 := exec.Command("echo", serviceOutputStr)
	c2 := exec.Command("awk", `{print $3}`)
	//fmt.Println(c1.String())
	//fmt.Println(c2.String())

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
	clusterIp := out.String()
	//fmt.Println("" + clusterIp)
	//  if err := c2.Run(); err != nil {
	//	log.Fatal(err)
	//  }
	return strings.TrimSpace(clusterIp)
}

// get the NodePort for the given service output and the clusterPort (internal K8s port)
// serviceOutput is the output of kubectl get service for a single service.
// TODO slice awk output on ','
func nodePortForClusterPort(serviceOutputStr string, clusterPort string) string {
	c1 := exec.Command("echo", serviceOutputStr)
	c2 := exec.Command("awk", `{print $5}`)
	//fmt.Println(c1.String())
	//fmt.Println(c2.String())

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
	// out contains  all nodeportsi, e.g. 9001:30589/TCP,9080:30151/TCP,8545:32119/TCP,8546:30510/TCP,30303:32238/TCP
	nodePortOutput := strings.TrimSpace(out.String())
	// example nodePort output: 9080:31973/TCP,8545:32734/TCP
	nodePorts := strings.Split(nodePortOutput, ",")
	//fmt.Println(fmt.Sprintf("nodePorts [%v]", nodePorts))
	// only one nodePort no filtering, just return the nodePort, e.g. cakeshop
	var nodePort string
	if len(nodePorts) == 1 {
		// get  8545:32734/TCP
		internalAndNodePort := strings.Split(nodePorts[0], ":")
		nodePortAndProtocol := strings.Split(internalAndNodePort[1], "/")
		nodePort = nodePortAndProtocol[0]
	} else if clusterPort != "" {
		for _, nodePortEntry := range nodePorts {
			internalAndNodePort := strings.Split(nodePortEntry, ":")
			_clusterPort := internalAndNodePort[0]
			// obtain the nodePort associated with the clusterIp
			if clusterPort == _clusterPort {
				nodePortAndProtocol := strings.Split(internalAndNodePort[1], "/")
				nodePort = nodePortAndProtocol[0]
			}
		}
	}

	return nodePort
}
func nodePortForService(serviceOutputStr string) string {
	return nodePortForClusterPort(serviceOutputStr, "")
}

func showPods(namespace string) {
	cmd := exec.Command("kubectl", "--namespace="+namespace, "get", "pods")
	var b bytes.Buffer
	cmd.Stdout = &b
	fmt.Print(b.String())
}

// https://www.reddit.com/r/golang/comments/2nd4pq/how_can_i_open_an_interactive_subprogram_from/
// runs a subcommand in interactive mode.
func dropIntoCmd(cmd *exec.Cmd) {
	//log.Printf(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runCmd(cmd *exec.Cmd) bytes.Buffer {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
