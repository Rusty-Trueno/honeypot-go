package controller

import (
	"honeypot/conf"
	"honeypot/core/transport"
)


func Run() {
	conf.Init()
	transport.Start(conf.GetConfig().Mqtt)
}

