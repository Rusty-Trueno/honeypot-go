package bypass

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"honeypot/util/constant"
)

type ReportResult struct {
	PotId    string `json:"potId"`
	Typex    string `json:"type"`
	SourceIp string `json:"sourceIp"`
	Info     string `json:"info"`
}

var (
	connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
		fmt.Printf("Connect succeed!\n")
	}

	connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
		fmt.Printf("Connect lost: %v\n", err)
	}

	Client MQTT.Client

	edgeNode = ""
)

func ReportToEdge(potId, typex, sourceIp, info string) {
	reportResult := ReportResult{
		PotId:    potId,
		Typex:    typex,
		SourceIp: sourceIp,
		Info:     info,
	}
	payload, _ := json.Marshal(reportResult)
	if token := Client.Publish(edgeNode+constant.BypassPotMsg, 0, false, payload); token.Wait() && token.Error() != nil {
		fmt.Errorf("publish pot msg failed, err is %v", token.Error())
	}
	return
}

func InitMqttClient(server, clientId, username, password, node string) error {
	edgeNode = node
	opts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientId).SetCleanSession(true)
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
		return fmt.Errorf("connect error: %v\n", token.Error())
	}
	return nil
}
