package kubeedge

import (
	"context"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"honeypot/core/honeypot"
	"honeypot/core/transport/mqtt"
	"honeypot/model"
	"honeypot/util/constant"
	"strings"
)

type Config struct {
	Server   string
	ClientID string
	Username string
	Password string
	Node     string
}
type Manager struct {
	Config
	Pots map[string]*honeypot.Honeypot
}

func New(cfg Config) (*Manager, error) {
	m := &Manager{
		Config: cfg,
		Pots:   make(map[string]*honeypot.Honeypot),
	}
	m.getAllPots(cfg.Node)
	return m, nil
}

func (m *Manager) Start() {
	ctx := context.TODO()
	go m.watchMemberUpdate(ctx.Done())
	m.watchAllPots()
	<-ctx.Done()
}

func (m *Manager) watchMemberUpdate(done <-chan struct{}) {
	var memberUpdate model.MemberUpdate
	if token := mqtt.Client.Subscribe(
		constant.NodeETPrefix+m.Node+constant.DeviceETMemberUpdated,
		2,
		func(client MQTT.Client, message MQTT.Message) {
			err := json.Unmarshal(message.Payload(), &memberUpdate)
			if err != nil {
				fmt.Errorf("json unmarshal err:%v\n", err)
			}
			fmt.Printf("update msg is %+v", memberUpdate)
			// handle added event
			for i := range memberUpdate.Added {
				device := memberUpdate.Added[i]
				pot := honeypot.NewPot(
					device.Id,
					*device.Twin["port"].Expected.Value,
					*device.Twin["protocol"].Expected.Value,
					*device.Twin["switch"].Expected.Value)
				m.Pots[device.Id] = pot
				go pot.Watch()
			}
			// handle removed event
			for i := range memberUpdate.Removed {
				device := memberUpdate.Removed[i]
				pot := m.Pots[device.Id]
				pot.UnWatch()
				delete(m.Pots, device.Id)
			}
		}); token.Wait() && token.Error() != nil {
		fmt.Errorf("get all pots failed, err is %v", token.Error())
	}
	<-done
}

func (m *Manager) watchAllPots() {
	for _, pot := range m.Pots {
		go pot.Watch()
	}
}

func (m *Manager) getAllPots(node string) {
	done := make(chan bool)
	var devices model.DeviceList
	if token := mqtt.Client.Subscribe(
		constant.NodeETPrefix+node+constant.DeviceETMemberResultSuffix,
		0,
		func(client MQTT.Client, message MQTT.Message) {
			err := json.Unmarshal(message.Payload(), &devices)
			if err != nil {
				fmt.Errorf("json unmarshal err:%v\n", err)
			}
			fmt.Printf("devices is %+v", devices)
			done <- true
		}); token.Wait() && token.Error() != nil {
		fmt.Errorf("get all pots failed, err is %v", token.Error())
	}
	payload, _ := json.Marshal(struct{}{})
	if token := mqtt.Client.Publish(constant.NodeETPrefix+node+constant.DeviceETMemberGetSuffix, 0, false, payload); token.Wait() && token.Error() != nil {
		fmt.Errorf("get all pots failed, err is %v", token.Error())
	}
	<-done
	for i := range devices.Devices {
		potId := devices.Devices[i].Id
		if strings.HasPrefix(potId, "pot") {
			device := getHoneypot(potId)
			pot := honeypot.NewPot(
				device.Id,
				*device.Twin["port"].Expected.Value,
				*device.Twin["protocol"].Expected.Value,
				*device.Twin["switch"].Expected.Value)
			m.Pots[potId] = pot
		}
	}
}

func getHoneypot(id string) *model.DeviceTwinUpdate {
	var device model.DeviceTwinUpdate
	done := make(chan bool)
	if token := mqtt.Client.Subscribe(
		constant.DeviceETPrefix+id+constant.TwinETGetResultSuffix,
		0,
		func(client MQTT.Client, message MQTT.Message) {
			err := json.Unmarshal(message.Payload(), &device)
			if err != nil {
				fmt.Errorf("json unmarshal err:%v\n", err)
			}
			fmt.Printf("\nhoneypot-%s is %+v", id, device)
			done <- true
		},
	); token.Wait() && token.Error() != nil {
		fmt.Errorf("get pot failed, err is %v", token.Error())
	}
	payload, _ := json.Marshal(struct{}{})
	if token := mqtt.Client.Publish(constant.DeviceETPrefix+id+constant.TwinETGetSuffix, 0, false, payload); token.Wait() && token.Error() != nil {
		fmt.Errorf("get pot failed, err is %v", token.Error())
	}
	<-done
	return &device
}
