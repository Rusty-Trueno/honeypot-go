package controller

import (
	"honeypot/conf"
	"honeypot/core/pool"
	"honeypot/core/proxy"
	"honeypot/core/transport/downstream"
	"honeypot/core/transport/upstream"
	"honeypot/model"
)

func Run() {
	conf.Init()
	wg, poolX := pool.New(1)
	defer poolX.Release()
	wg.Add(1)
	result := make(chan model.FuncResult)
	forWardJob := new(proxy.ForWardJob)
	forWardJob.ClientMap = make(map[string]*proxy.ForWardClient, 500)
	forWardJob.Config = &model.ForwardConfig{
		Name:     "redis",
		SrcAddr:  "127.0.0.1",
		SrcPort:  6000,
		DestAddr: "127.0.0.1",
		DestPort: 6378,
		Protocol: "TCP",
	}
	forWardJob.UdpForwardJob = proxy.NewUdpForward()
	go forWardJob.StartJob(result)
	//go forWardJob.HandleUnMatch(status.GetRedisUnMatch())
	downstream.ClientInit(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.DownClientId, "", "")
	upstream.ClientInit(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.UpClientId, "", "")
	wg.Wait()
}
