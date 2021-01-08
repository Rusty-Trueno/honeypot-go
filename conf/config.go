package conf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var DEV_CONFIG_FILE_PATH = "conf/config.dev.yaml"

type MqttConfig struct {
	Server		string 		`yaml:"server"`
	Mode		int64		`yaml:"mode"`
	ClientId	string		`yaml:"clientId"`
}

type ConfigFile struct {
	Mqtt 		MqttConfig		`yaml:"mqtt"`
}

var config *ConfigFile

func Init() {
	yamlFile, err := ioutil.ReadFile(DEV_CONFIG_FILE_PATH)
	if err != nil {
		log.Fatalf("io error: %v\n", err)
	}


	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("json unmarshal error: %v\n", err)
	}
}

func GetConfig() *ConfigFile {
	return config
}
