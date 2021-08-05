package bypass

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"honeypot/core/event"
	"honeypot/util/constant"
)

var (
	connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
		fmt.Printf("Connect succeed!\n")
	}

	connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
		fmt.Printf("Connect lost: %v\n", err)
	}

	Client MQTT.Client

	edgeNode = ""

	Bk Backend
)

func InitMqttClient(server, clientId, username, password, node string) error {
	edgeNode = node
	ch := make(chan map[string]interface{}, 100)
	Bk = Backend{
		ch: ch,
	}
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
	go Bk.run()
	return nil
}

type Backend struct {
	ch chan map[string]interface{}
}

func (b Backend) run() {
	for e := range b.ch {
		payload, err := json.Marshal(e)
		if err != nil {
			fmt.Errorf("json marshal failed, error is %v", err)
		}
		fmt.Printf("payload is %s", e)
		if token := Client.Publish(edgeNode+constant.BypassPotMsg, 0, false, payload); token.Wait() && token.Error() != nil {
			fmt.Errorf("publish pot msg failed, err is %v", token.Error())
		}
	}
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (b Backend) Send(e event.Event) {
	mp := make(map[string]interface{})

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	b.ch <- mp
}
