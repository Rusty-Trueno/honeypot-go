package controller

import (
	"fmt"
	"honeypot/conf"
	"honeypot/core/pool"
	"honeypot/core/transport/kubeedge"
	"honeypot/core/transport/mqtt"
)

func Run(node string) {
	conf.Init()
	wg, poolX := pool.New(1)
	defer poolX.Release()
	wg.Add(1)
	err := mqtt.InitMqttClient(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.DownClientId, "", "")
	if err != nil {
		fmt.Errorf("init mqtt client failed, err is %v", err)
	}
	cfg := kubeedge.Config{
		Server:   conf.GetConfig().Mqtt.Server,
		ClientID: conf.GetConfig().Mqtt.DownClientId,
		Node:     node,
	}
	m, err := kubeedge.New(cfg)
	if err != nil {
		fmt.Errorf("new kubeedge failed, err is %v", err)
	}
	fmt.Printf("\nnew kubeedge succeed, manager is %+v", m)
	go m.Start()
	wg.Wait()
}
