package conf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var DEV_CONFIG_FILE_PATH = "conf/config.dev.yaml"

type MqttConfig struct {
	Server        string `yaml:"server"`
	Mode          int64  `yaml:"mode"`
	DownClientId  string `yaml:"downClientId"`
	UpClientId    string `yaml:"upClientId"`
	Kdd99ClientId string `yaml:"kdd99ClientId"`
}

type HoneypotConfig struct {
	RedisConfig  RedisConfig  `yaml:"redis,omitempty"`
	MysqlConfig  MysqlConfig  `yaml:"mysql,omitempty"`
	TelnetConfig TelnetConfig `yaml:"telnet,omitempty"`
	WebConfig    WebConfig    `yaml:"web,omitempty"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type MysqlConfig struct {
	Addr  string `yaml:"addr"`
	Files string `yaml:"files"`
}

type TelnetConfig struct {
	Addr string `yaml:"addr"`
}

type WebConfig struct {
	Addr     string `yaml:"addr"`
	Template string `yaml:"template"`
	Index    string `yaml:"index"`
	Static   string `yaml:"static"`
	Url      string `yaml:"url"`
}

type ConfigFile struct {
	Mqtt           MqttConfig     `yaml:"mqtt"`
	HoneypotConfig HoneypotConfig `yaml:"honeypot"`
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
