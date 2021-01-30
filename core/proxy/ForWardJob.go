package proxy

import (
	"fmt"
	"honeypot/constant"
	"honeypot/model"
	"honeypot/util"

	"net"
	"sync"
	"time"
)

type ForWardJob struct {
	Config        *model.ForwardConfig
	ClientMap     map[string]*ForWardClient
	ClientMapLock sync.Mutex
	Status        byte
	PortListener  net.Listener
	UdpForwardJob *UdpForward
}

func (_self *ForWardJob) StartJob(result chan model.FuncResult) {

	sourceAddr := fmt.Sprint(_self.Config.SrcAddr, ":", _self.Config.SrcPort)
	destAddr := fmt.Sprint(_self.Config.DestAddr, ":", _self.Config.DestPort)

	resultData := &model.FuncResult{Code: 0, Msg: ""}
	var err error
	if _self.IsUdpJob() {
		//_self.PortListener, err = NetUtils.NewKCP(sourceAddr, Common.DefaultKcpSetting())
		//_self.UdpForwardJob.UdpListenerConn, err = NetUtils.NewUDP(sourceAddr)

		err = _self.UdpForwardJob.DoUdpForward(sourceAddr, destAddr)

		if err != nil {
			fmt.Printf("启动UDP监听 ", sourceAddr, " 出错：", err)
			resultData.Code = 1
			resultData.Msg = fmt.Sprint("启动UDP监听 ", sourceAddr, " 出错：", err)
			result <- *resultData
			return
		}

		_self.Status = constant.RunStatus_Running
		fmt.Printf("启动UDP端口转发，从 ", sourceAddr, " 到 ", destAddr)
		result <- *resultData

	} else {
		_self.PortListener, err = util.NewTCP(sourceAddr)

		if err != nil {
			fmt.Printf("启动监听 ", sourceAddr, " 出错：", err)
			resultData.Code = 1
			resultData.Msg = fmt.Sprint("启动监听 ", sourceAddr, " 出错：", err)
			result <- *resultData
			return
		}

		_self.Status = constant.RunStatus_Running
		fmt.Printf("启动端口转发，从 %v, 到 %v", sourceAddr, destAddr)

		_self.doTcpForward(destAddr)

	}

}

func (_self *ForWardJob) doTcpForward(destAddr string) {

	for {
		realClientConn, err := _self.PortListener.Accept()
		if err != nil {
			fmt.Printf("Forward Accept err:", err.Error())
			fmt.Printf(fmt.Sprint("转发出现异常：", _self.Config.SrcAddr, ":", _self.Config.SrcPort, "->", destAddr))
			_self.StopJob()
			break
		}

		var destConn net.Conn
		if _self.Config.Protocol == "UDP" {
			//destConn, err = Common.DialKcpTimeout(destAddr, 100)
			destConn, err = net.DialTimeout("UDP", destAddr, 30*time.Second)
		} else {
			destConn, err = net.DialTimeout("tcp", destAddr, 30*time.Second)
		}

		if err != nil {

			//break
			continue

		}

		forwardClient := &ForWardClient{realClientConn, destConn, _self.ClosedCallBack}
		go forwardClient.StartForward()

		_self.RegistryClient(_self.GetClientId(realClientConn), forwardClient)
		//_self.RegistryClient(fmt.Sprint(sourceAddr, "_", "TCP", "_", id), forwardClient)

	}
}

func (_self *ForWardJob) ClosedCallBack(srcConn net.Conn, destConn net.Conn) {

	_self.UnRegistryClient(_self.GetClientId(srcConn))
}

func (_self *ForWardJob) GetClientId(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

func (_self *ForWardJob) RegistryClient(srcAddr string, forwardClient *ForWardClient) {
	_self.ClientMapLock.Lock()
	defer _self.ClientMapLock.Unlock()

	_self.ClientMap[srcAddr] = forwardClient

}

func (_self *ForWardJob) UnRegistryClient(srcAddr string) {
	_self.ClientMapLock.Lock()
	defer _self.ClientMapLock.Unlock()

	delete(_self.ClientMap, srcAddr)

}

func (_self *ForWardJob) IsJobRunning() bool {

	return _self.Status == constant.RunStatus_Running

}

func (_self *ForWardJob) IsUdpJob() bool {
	return util.ToUpper(_self.Config.Protocol) == "UDP"
}

func (_self *ForWardJob) StopJob() {

	if _self.IsUdpJob() {
		_self.stopUdpJob()
	} else {
		_self.stopTcpJob()
	}

	_self.Status = constant.RunStatus_Stoped
}

func (_self *ForWardJob) stopTcpJob() {

	_self.PortListener.Close()

	for _, client := range _self.ClientMap {
		client.StopForward()
	}

	_self.ClientMap = nil
}

func (_self *ForWardJob) stopUdpJob() {

	_self.UdpForwardJob.Close()
}
