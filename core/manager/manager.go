package manager

import (
	"fmt"
	"github.com/panjf2000/ants"
	"honeypot/core/pool"
	"honeypot/core/proxy"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	ForwardMap map[string]*proxy.ForWardJob
}

var wg sync.WaitGroup

var poolX *ants.Pool

func (_self *Manager) StartManager(addr string) error {
	_self.ForwardMap = make(map[string]*proxy.ForWardJob)
	// 建立socket，监听端口
	netListen, _ := net.Listen("tcp", addr)

	defer netListen.Close()

	wg, poolX = pool.New(1)
	defer poolX.Release()

	for {
		wg.Add(1)
		poolX.Submit(func() {
			time.Sleep(time.Second * 2)

			conn, err := netListen.Accept()

			if err != nil {
				fmt.Printf("manager连接失败， error is %v\n", err)
				wg.Done()
				return
			}

			fmt.Printf("Manager 连接成功！\n")

			go _self.handleConnection(conn)

			wg.Done()
		})
	}
}

func (_self *Manager) handleConnection(conn net.Conn) {
	fmt.Printf("new connection\n")
	//redisConn := _self.ForwardMap["redis"].
	//go func() {
	//	_, err := io.Copy(_self.DestConn, conn)
	//	if err != nil {
	//		fmt.Printf("客户端来源数据转发到目标端口异常：%v", err)
	//	}
	//}()
	//
	//go func() {
	//	_, err := io.Copy(_self.SrcConn, _self.DestConn)
	//	if err != nil {
	//		fmt.Printf("目标端口返回响应数据异常：%v", err)
	//	}
	//}()
}

func getRemotePort(conn net.Conn) (int, error) {
	strs := strings.Split(conn.RemoteAddr().String(), ":")
	port, err := strconv.Atoi(strs[1])
	return port, err
}
