package honeypot

import (
	"context"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"honeypot/conf"
	"honeypot/core/protocol/mysql"
	"honeypot/core/protocol/redis"
	"honeypot/core/protocol/telnet"
	"honeypot/core/protocol/web"
	"honeypot/core/transport/mqtt"
	"honeypot/model"
	"honeypot/util/constant"
	"honeypot/util/linux"
	"time"
)

type Honeypot struct {
	Name     string
	StopCh   chan bool
	ctx      context.Context
	cancel   context.CancelFunc
	Port     string
	Protocol string
	Switch   string
}

func NewPot(Name, Port, Protocol, Switch string) *Honeypot {
	ctx, cancel := context.WithCancel(context.Background())
	return &Honeypot{
		Name:     Name,
		Protocol: Protocol,
		Port:     Port,
		Switch:   Switch,
		StopCh:   make(chan bool),
		cancel:   cancel,
		ctx:      ctx,
	}
}

func (h *Honeypot) start() {
	addr := ""
	switch h.Protocol {
	case constant.Redis:
		if h.Port == "" {
			addr = conf.GetConfig().HoneypotConfig.RedisConfig.Addr
		} else {
			addr = fmt.Sprintf("0.0.0.0:%s", h.Port)
		}
		go redis.Start(h.Name, addr, h.StopCh)
	case constant.Mysql:
		if h.Port == "" {
			addr = conf.GetConfig().HoneypotConfig.MysqlConfig.Addr
		} else {
			addr = fmt.Sprintf("0.0.0.0:%s", h.Port)
		}
		go mysql.Start(h.Name, addr, conf.GetConfig().HoneypotConfig.MysqlConfig.Files, h.StopCh)
	case constant.Telnet:
		if h.Port == "" {
			addr = conf.GetConfig().HoneypotConfig.TelnetConfig.Addr
		} else {
			addr = fmt.Sprintf("0.0.0.0:%s", h.Port)
		}
		go telnet.Start(h.Name, addr, h.StopCh)
	case constant.Web:
		if h.Port == "" {
			addr = conf.GetConfig().HoneypotConfig.WebConfig.Addr
		} else {
			addr = fmt.Sprintf("0.0.0.0:%s", h.Port)
		}
		webConfig := conf.GetConfig().HoneypotConfig.WebConfig
		go web.Start(addr, webConfig.Template, webConfig.Static, webConfig.Url, webConfig.Index, h.StopCh)
	default:
		fmt.Printf("unknown protocal: %s\n", h.Protocol)
	}
}

func (h *Honeypot) Watch() {
	err := h.updatePort(h.Port)
	if err != nil {
		fmt.Errorf("update port failed, err is %v", err)
	}
	err = h.syncProtocol(h.Protocol)
	if err != nil {
		fmt.Errorf("sync protocol failed, err is %v", err)
	}
	go h.watchPotStatus()
	if h.Switch == "ON" {
		h.start()
	}
	if token := mqtt.Client.Subscribe(
		constant.DeviceETPrefix+h.Name+constant.TwinETDeltaSuffix,
		0,
		func(client MQTT.Client, message MQTT.Message) {
			var device model.DeviceTwinUpdate
			err := json.Unmarshal(message.Payload(), &device)
			if err != nil {
				fmt.Errorf("json unmarshal err:%v\n", err)
			}
			fmt.Printf("\nsync twins -> un sync twins is %+v", device.Delta)
			for k, v := range device.Delta {
				switch k {
				case "switch":
					if h.Switch != v {
						if v == "ON" {
							h.start()
						} else if v == "OFF" {
							h.stop()
						}
					}
				case "port":
					if h.Port != v {
						h.changeBindPort(v)
					}
					err := h.updatePort(v)
					if err != nil {
						fmt.Errorf("update port failed, err is %v", err)
					}
				case "protocol":
					err := h.syncProtocol(v)
					if err != nil {
						fmt.Errorf("sync protocol failed, err is %v", err)
					}
				}
			}

		}); token.Wait() && token.Error() != nil {
		fmt.Errorf("watch pot failed, err is %v", token.Error())
	}
	<-h.ctx.Done()
	fmt.Println("stop watch twins")
}

func (h *Honeypot) UnWatch() {
	if h.Switch == "ON" {
		h.stop()
	}
	h.cancel()
}

func (h *Honeypot) stop() {
	h.StopCh <- true
}

func (h *Honeypot) changeBindPort(newPort string) {
	h.Port = newPort
	if h.Switch == "ON" {
		h.stop()
		h.start()
	}
}

func (h *Honeypot) watchPotStatus() {
	for {
		select {
		case <-time.After(10 * time.Second):
			var status string
			if !linux.CheckPort(h.Port) {
				// 如果端口不再监听了
				fmt.Println("not listen")
				status = "OFF"
			} else {
				fmt.Println("listening")
				status = "ON"
			}
			err := h.updateSwitch(status)
			if err != nil {
				fmt.Errorf("update switch failed, err is %v", err)
			}
		case <-h.ctx.Done():
			fmt.Println("stop watch pod status")
			return
		}
	}
}

func (h *Honeypot) updateSwitch(status string) error {
	h.Switch = status
	switchTwin := model.MsgTwin{
		Actual: &model.TwinValue{
			Value: &status,
		},
	}
	updateDevice := model.DeviceTwinUpdate{
		Twin: map[string]*model.MsgTwin{
			"switch": &switchTwin,
		},
	}
	payload, err := json.Marshal(updateDevice)
	if err != nil {
		return err
	}
	if token := mqtt.Client.Publish(constant.DeviceETPrefix+h.Name+constant.TwinETUpdateSuffix, 0, false, payload); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

func (h *Honeypot) updatePort(port string) error {
	portTwin := model.MsgTwin{
		Actual: &model.TwinValue{
			Value: &port,
		},
	}
	updateDevice := model.DeviceTwinUpdate{
		Twin: map[string]*model.MsgTwin{
			"port": &portTwin,
		},
	}
	payload, err := json.Marshal(updateDevice)
	if err != nil {
		return err
	}
	if token := mqtt.Client.Publish(constant.DeviceETPrefix+h.Name+constant.TwinETUpdateSuffix, 0, false, payload); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}

func (h *Honeypot) syncProtocol(protocol string) error {
	protocolTwin := model.MsgTwin{
		Actual: &model.TwinValue{
			Value: &protocol,
		},
	}
	updateDevice := model.DeviceTwinUpdate{
		Twin: map[string]*model.MsgTwin{
			"protocol": &protocolTwin,
		},
	}
	payload, err := json.Marshal(updateDevice)
	if err != nil {
		return err
	}
	if token := mqtt.Client.Publish(constant.DeviceETPrefix+h.Name+constant.TwinETUpdateSuffix, 0, false, payload); token.Wait() && token.Error() != nil {
		return err
	}
	return nil
}
