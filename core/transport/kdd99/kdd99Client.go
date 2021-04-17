package kdd99

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var Kdd99Client MQTT.Client

var kdd99ConnectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Printf("Connect succeed!\n")
}

var kdd99ConnectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

var kdd99MsgHandler MQTT.MessageHandler = func(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("msg is: %s\n", message.Payload())

}

func Kdd99ClientStart(server, clientID, username, password string) {
	opts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientID).SetCleanSession(true)
	if username != "" {
		opts.SetUsername(username)
		if password != "" {
			opts.SetPassword(password)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	opts.OnConnect = kdd99ConnectHandler
	opts.OnConnectionLost = kdd99ConnectLostHandler
	opts.AutoReconnect = true
	Kdd99Client = MQTT.NewClient(opts)
	if token := Kdd99Client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Errorf("connect error: %v\n", token.Error())
	}
	token := Kdd99Client.Subscribe("kdd99", 0, kdd99MsgHandler)
	if token.Wait() && token.Error() != nil {
		fmt.Errorf("subscribe error: %v\n", token.Error())
	}
}

func Kdd99ClientStop() {
	Kdd99Client.Disconnect(0)
	fmt.Printf("Kdd99Client stoped")
}
