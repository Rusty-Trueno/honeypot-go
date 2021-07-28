package mqtt

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
		fmt.Printf("Connect succeed!\n")
	}

	connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
		fmt.Printf("Connect lost: %v\n", err)
	}

	Client MQTT.Client
)

func InitMqttClient(server, clientId, username, password string) error {
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
