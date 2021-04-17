package status

import (
	"fmt"
	"honeypot/conf"
	"honeypot/core/transport/kdd99"
)

var connCnt = 0

var connIn = make(chan bool)

var connOut = make(chan bool)

func CheckConn() {
	for {
		select {
		case _ = <-connIn:
			connCnt++
			fmt.Printf("conn in, connCnt is %d\n", connCnt)
			kdd99.Kdd99ClientStart(conf.GetConfig().Mqtt.Server, conf.GetConfig().Mqtt.Kdd99ClientId, "", "")
		case _ = <-connOut:
			if connCnt > 0 {
				connCnt--
				fmt.Printf("conn out, connCnt is %d\n", connCnt)
				if connCnt == 0 {
					kdd99.Kdd99ClientStop()
				}
			}
		}
	}
}

func GetConnIn() chan bool {
	return connIn
}

func SetConnIn() {
	connIn <- true
}

func GetConnOut() chan bool {
	return connOut
}

func SetConnOut() {
	connOut <- true
}
