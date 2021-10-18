package controller

import (
	"fmt"
	"honeypot/conf"
	"honeypot/core/db/tdengine"
	"honeypot/core/pool"
	"honeypot/core/pushers/bypass"
	"honeypot/core/pushers/timeseries"
	"honeypot/core/storage"
	"honeypot/core/transport/kubeedge"
	"honeypot/core/transport/mqtt"
)

func Run(node, env string) {
	conf.Init()
	wg, poolX := pool.New(1)
	defer poolX.Release()
	wg.Add(1)
	err := tdengine.Setup(node)
	if err != nil {
		fmt.Errorf("init tdengine failed, error is %v", err)
	}
	storage.SetDataDir("./")
	err = mqtt.InitMqttClient(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.DownClientId, "", "")
	if err != nil {
		fmt.Errorf("init mqtt client failed, err is %v", err)
	}
	err = bypass.InitMqttClient(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.UpClientId, "", "", node)
	if err != nil {
		fmt.Errorf("init bypass mqtt client failed, err is %v", err)
	}
	err = timeseries.New()
	if err != nil {
		fmt.Errorf("init timeseries failed, err is %v", err)
	}
	cfg := kubeedge.Config{
		Server:   conf.GetConfig().Mqtt.Server,
		ClientID: conf.GetConfig().Mqtt.DownClientId,
		Node:     node,
		Env:      env,
	}
	m, err := kubeedge.New(cfg)
	if err != nil {
		fmt.Errorf("new kubeedge failed, err is %v", err)
	}
	fmt.Printf("\nnew kubeedge succeed, manager is %+v", m)
	go m.Start()
	wg.Wait()
}
