package controller

import (
	"honeypot/conf"
	"honeypot/core/pool"
	"honeypot/core/status"
	"honeypot/core/transport/downstream"
	"honeypot/core/transport/upstream"
)

func Run() {
	conf.Init()
	wg, poolX := pool.New(1)
	defer poolX.Release()
	wg.Add(1)
	downstream.ClientInit(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.DownClientId, "", "")
	upstream.ClientInit(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.UpClientId, "", "")
	go status.CheckConn()
	wg.Wait()
}
