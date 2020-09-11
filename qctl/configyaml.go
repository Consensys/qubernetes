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

type Quorum struct {
	Consensus     string `yaml:"consensus"`
	QuorumVersion string `yaml:"Quorum_Version"`
}

type Tm struct {
	Name      string `yaml:"Name"`
	TmVersion string `yaml:"Tm_Version"`
}

type NodeEntry struct {
	NodeUserIdent string      `yaml:"Node_UserIdent"`
	KeyDir        string      `yaml:"Key_Dir"`
	QuorumEntry   QuorumEntry `yaml:"quorum"`
}

type QConfig struct {
	Genesis struct {
		Consensus     string `yaml:"consensus"`
		QuorumVersion string `yaml:"Quorum_Version"`
		Chain_Id      string `yaml:"Chain_Id"`
	}

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
