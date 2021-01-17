package upstream

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var Client MQTT.Client

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Printf("Connect succeed!\n")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
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
}

func Publish(topic string, msg interface{}) {
	fmt.Printf("Publis mqtt: %v, topic is %s\n", msg, topic)
	token := Client.Publish(topic, 0, false, msg)
	if token.Wait() && token.Error() != nil {
		fmt.Errorf("publish error: %v\n", token.Error())
	}
}
