package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var (
	qubeTemplateYaml = `
    genesis:
      # supported: (raft | istanbul)
      consensus: istanbul
      Quorum_Version: 2.6.0
      Chain_Id: 1000
    nodes:
    `
)

type QuorumEntry struct {
	Quorum Quorum
	Tm     Tm
}

type GethEntry struct {
	GetStartupParams string `yaml:"Geth_Startup_Params"` //--raftjoinexisting 7
}

type Quorum struct {
	Consensus      string `yaml:"consensus"`
	QuorumVersion  string `yaml:"Quorum_Version"`
	DockerRepoFull string `yaml:"Docker_Repo_Full"` //quorum-local:latest
}

type Tm struct {
	Name      string `yaml:"Name"`
	TmVersion string `yaml:"Tm_Version"`
}

type NodeEntry struct {
	NodeUserIdent string      `yaml:"Node_UserIdent"`
	KeyDir        string      `yaml:"Key_Dir"`
	QuorumEntry   QuorumEntry `yaml:"quorum"`
	GethEntry     GethEntry   `yaml:"geth"`
}

type Prometheus struct {
	//#monitor_params_geth: --metrics --metrics.expensive --pprof --pprofaddr=0.0.0.0
	//monitorParamsGeth string `yaml:"monitor_params_geth"`
	NodePort string `yaml:"nodePort_prom,omitempty"`
	Enabled  bool   `yaml:"enabled,omitempty"`
}

type Cakeshop struct {
	Version string `yaml:"version,omitempty"`
	Service struct {
		Type     string `yaml:"type,omitempty"`
		NodePort string `yaml:"nodePort,omitempty"`
	}
}

type Ingress struct {
	//OneToMany | OneToOne
	Strategy string `yaml:"Strategy,omitempty"`
	Host     string `yaml:"Host,omitempty"`
}

type K8s struct {
	Service struct {
		Type    string  `yaml:"type,omitempty"`
		Ingress Ingress `yaml:"Ingress,omitempty"`
	}
}

type QConfig struct {
	Genesis struct {
		Consensus     string `yaml:"consensus"`
		QuorumVersion string `yaml:"Quorum_Version"`
		Chain_Id      string `yaml:"Chain_Id"`
	}
	Prometheus Prometheus `yaml:"prometheus,omitempty"`
	Cakeshop   Cakeshop   `yaml:"cakeshop,omitempty"`
	K8s        K8s        `yaml:"k8s,omitempty"`

	Nodes []NodeEntry
}

func GetYamlConfig() QConfig {
	config := QConfig{}
	err := yaml.Unmarshal([]byte(qubeTemplateYaml), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}

func LoadYamlConfig(filename string) (QConfig, error) {
	config := QConfig{}
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading file: %v err: %v", filename, err)
		return config, err
	}

	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Fatalf("error unmarshalling file: %v err: %v", filename, err)
		return config, err
	}
	return config, nil
}

func WriteYamlConfig(qconfig QConfig, filename string) (QConfig, error) {
	bs, err := yaml.Marshal(&qconfig)
	if err != nil {
		log.Fatalf("error: %v", err)
		return qconfig, err
	}
	ioutil.WriteFile(filename, bs, os.ModePerm)
	return qconfig, err
}

func (q QConfig) ToString() string {
	d, err := yaml.Marshal(&q)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return string(d)
}
