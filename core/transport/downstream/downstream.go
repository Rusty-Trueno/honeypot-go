package downstream

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"honeypot/conf"
	"honeypot/core/protocol/mysql"
	"honeypot/core/protocol/redis"
	"honeypot/core/protocol/telnet"
	"honeypot/core/status"
)

var Client MQTT.Client

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Printf("Connect succeed!\n")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

var msgHandler MQTT.MessageHandler = func(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("msg is: %s\n", message.Payload())
	var order Order
	err := json.Unmarshal(message.Payload(), &order)
	if err != nil {
		fmt.Errorf("json unmarshal err:%v\n", err)
	}
	if order.Target == "redis" {
		if order.Move == "open" {
			if !status.GetRedisStatus() {
				go redis.Start(conf.GetConfig().HoneypotConfig.RedisConfig.Addr, status.GetRedisDone())
				status.SetRedisStatus(true)
			}
		} else if order.Move == "stop" {
			if status.GetRedisStatus() {
				status.SetRedisDone(true)
				status.SetRedisStatus(false)
			}
		}
	} else if order.Target == "mysql" {
		if order.Move == "open" {
			if !status.GetMysqlStatus() {
				go mysql.Start(conf.GetConfig().HoneypotConfig.MysqlConfig.Addr, conf.GetConfig().HoneypotConfig.MysqlConfig.Files, status.GetMysqlDone())
				status.SetMysqlStatus(true)
			}
		} else if order.Move == "stop" {
			if status.GetMysqlStatus() {
				status.SetMysqlDone(true)
				status.SetMysqlStatus(false)
			}
		}
	} else if order.Target == "telnet" {
		if order.Move == "open" {
			if !status.GetTelnetStatus() {
				go telnet.Start(conf.GetConfig().HoneypotConfig.TelnetConfig.Addr, status.GetTelnetDone())
				status.SetTelnetStatus(true)
			}
		} else if order.Move == "stop" {
			if status.GetTelnetStatus() {
				status.SetTelnetDone(true)
				status.SetTelnetStatus(false)
			}
		}
	}
}

type Order struct {
	Move   string `json:"move"`
	Target string `json:"target"`
}

func ClientInit(server, clientID, username, password string) {
	opts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientID).SetCleanSession(true)
	if username != "" {
		opts.SetUsername(username)
		if password != "" {
			opts.SetPassword(password)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true
	Client = MQTT.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Errorf("connect error: %v\n", token.Error())
	}
	token := Client.Subscribe("CloudOrder", 0, msgHandler)
	if token.Wait() && token.Error() != nil {
		fmt.Errorf("subscribe error: %v\n", token.Error())
	}
}
