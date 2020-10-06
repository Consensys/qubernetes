package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var (
	acceptanceTestTemplateYaml = `
    quorum:
      nodes:
        Node1:
          privacy-address: 4qceb9wVWDrYXqwgmu23clRQDHZeFTP0xDfEpMzm7yg=
          url: http://192.168.64.49:30102
          third-party-url: http://192.168.64.49:32614
    `
)

type AcceptTestConfig struct {
	Quorum struct {
		Nodes []ATNodeEntry
	}
}

type ATNodeEntry struct {
	TmPublicKey string `yaml:"privacy-address"`
	GethURL     string `yaml:"url"`
	TmURL       string `yaml:"third-party-url"`
}

func LoadAcTYamlConfig(filename string) (AcceptTestConfig, error) {
	config := AcceptTestConfig{}
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

func WriteAcTYamlConfig(config AcceptTestConfig, filename string) (AcceptTestConfig, error) {
	bs, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("error: %v", err)
		return config, err
	}
	ioutil.WriteFile(filename, bs, os.ModePerm)
	return config, err
}

func (q AcceptTestConfig) ToString() string {
	d, err := yaml.Marshal(&q)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return string(d)
}
