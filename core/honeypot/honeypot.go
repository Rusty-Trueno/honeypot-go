package honeypot

import (
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
)

type Honeypot struct {
	Name     string
	Protocol string
	StopCh   chan bool
	DoneCh   chan bool
	Twin     map[string]*model.MsgTwin
}

func (h *Honeypot) start() {
	switch h.Protocol {
	case constant.Redis:
		go redis.Start(conf.GetConfig().HoneypotConfig.RedisConfig.Addr, h.StopCh)
	case constant.Mysql:
		go mysql.Start(conf.GetConfig().HoneypotConfig.MysqlConfig.Addr, conf.GetConfig().HoneypotConfig.MysqlConfig.Files, h.StopCh)
	case constant.Telnet:
		go telnet.Start(conf.GetConfig().HoneypotConfig.TelnetConfig.Addr, h.StopCh)
	case constant.Web:
		webConfig := conf.GetConfig().HoneypotConfig.WebConfig
		go web.Start(webConfig.Addr, webConfig.Template, webConfig.Static, webConfig.Url, webConfig.Index, h.StopCh)
	default:
		fmt.Printf("unknown protocal: %s\n", h.Protocol)
	}
}

func (h *Honeypot) Watch() {
	if *h.Twin["switch"].Expected.Value == "ON" {
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
			fmt.Printf("\ncurrent device is %+v", device)
			h.Twin = device.Twin
			if *device.Twin["switch"].Expected.Value == "OFF" {
				h.stop()
			} else if *device.Twin["switch"].Expected.Value == "ON" {
				h.start()
			}
		}); token.Wait() && token.Error() != nil {
		fmt.Errorf("watch pot failed, err is %v", token.Error())
	}
	<-h.DoneCh
}

func (h *Honeypot) UnWatch() {
	h.DoneCh <- true
}

func (h *Honeypot) stop() {
	h.StopCh <- true
}
