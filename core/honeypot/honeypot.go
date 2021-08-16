package honeypot

import (
	"context"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"honeypot/core/listener"
	"honeypot/core/pushers"
	bypass2 "honeypot/core/pushers/bypass"
	"honeypot/core/pushers/eventbus"
	"honeypot/core/services"
	"honeypot/core/transport/mqtt"
	"honeypot/model"
	"honeypot/util/constant"
	"honeypot/util/linux"
	"honeypot/util/windows"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "honeypot/core/listener/socket"

	_ "honeypot/core/pushers/console"
	_ "honeypot/core/pushers/eventbus"

	_ "honeypot/core/services/bannerfmt"
	_ "honeypot/core/services/decoder"
	_ "honeypot/core/services/docker"
	_ "honeypot/core/services/elasticsearch"
	_ "honeypot/core/services/eos"
	_ "honeypot/core/services/ethereum"
	_ "honeypot/core/services/filesystem"
	_ "honeypot/core/services/ftp"
	_ "honeypot/core/services/ipp"
	_ "honeypot/core/services/ja3/crypto/tls"
	_ "honeypot/core/services/ldap"
	_ "honeypot/core/services/redis"
	_ "honeypot/core/services/smtp"
	_ "honeypot/core/services/ssh"
	_ "honeypot/core/services/telnet"
	_ "honeypot/core/services/vnc"
)

type Honeypot struct {
	Name     string
	StopCh   chan bool
	ctx      context.Context
	cancel   context.CancelFunc
	Port     string
	Protocol string
	Switch   string
	Env      string
}

func NewPot(Name, Port, Protocol, Switch, Env string) *Honeypot {
	ctx, cancel := context.WithCancel(context.Background())
	return &Honeypot{
		Name:     Name,
		Protocol: Protocol,
		Port:     Port,
		Switch:   Switch,
		StopCh:   make(chan bool),
		cancel:   cancel,
		ctx:      ctx,
		Env:      Env,
	}
}

func (h *Honeypot) start() {
	go h.launchPot()
}

func (h *Honeypot) Watch() {
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
			if h.Env == constant.Windows {
				if !windows.CheckPort(h.Port) {
					// 如果端口不再监听了
					fmt.Println("not listen")
					status = "OFF"
				} else {
					fmt.Println("listening")
					status = "ON"
				}
			} else if h.Env == constant.Linux {
				if !linux.CheckPort(h.Port) {
					// 如果端口不再监听了
					fmt.Println("not listen")
					status = "OFF"
				} else {
					fmt.Println("listening")
					status = "ON"
				}
			}
			fmt.Printf("status is %s\n", status)
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

func (h *Honeypot) launchPot() {
	// init event bus
	bc := pushers.NewBusChannel()
	bus := eventbus.New()
	bus.Subscribe(bc)
	bus.Subscribe(bypass2.Bk)
	// init service
	fn, ok := services.Get(h.Protocol)
	if !ok {
		fmt.Errorf(color.RedString("Could not find service %s", h.Protocol))
		return
	}
	options := []services.ServicerFunc{
		services.WithChannel(bus),
	}
	service := fn(options...)
	// init listener
	listenerFunc, ok := listener.Get("socket")
	if !ok {
		fmt.Errorf(color.RedString("Listener not support socket type"))
		return
	}

	l, err := listenerFunc(
		listener.WithChannel(bus),
	)
	if err != nil {
		fmt.Errorf("Error init listener")
		return
	}

	a, ok := l.(listener.AddAddresser)
	if !ok {
		fmt.Errorf("Listener error")
		return
	}
	addr, _, _, err := ToAddr(fmt.Sprintf("tcp/%s", h.Port))
	if err != nil {
		fmt.Errorf("Error parsing port string: %s", err.Error())
		return
	}
	if addr == nil {
		fmt.Errorf("Failed to bind: addr is nil")
		return
	}
	a.AddAddress(addr)

	//start listen
	ctx, cancel := context.WithCancel(context.Background())
	if err := l.Start(ctx); err != nil {
		fmt.Errorf(color.RedString("Error starting listener: %s", err.Error()))
		return
	}

	incoming := make(chan net.Conn)
	var conn net.Conn
	go func() {
		for {
			conn, err = l.Accept()
			if err != nil {
				panic(err)
			}

			incoming <- conn

			// in case of goroutine starvation
			// with many connection and single procs
			runtime.Gosched()
		}
	}()

	go func() {
		<-h.StopCh
		cancel()
		err := l.Stop()
		if err != nil {
			fmt.Errorf("close socker failed, err is %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-incoming:
			go service.Handle(ctx, conn)
		}
	}

}

// Addr, proto, port, error
func ToAddr(input string) (net.Addr, string, int, error) {
	parts := strings.Split(input, "/")

	if len(parts) != 2 {
		return nil, "", 0, fmt.Errorf("wrong format (needs to be \"protocol/(host:)port\")")
	}

	proto := parts[0]

	host, port, err := net.SplitHostPort(parts[1])
	if err != nil {
		port = parts[1]
	}

	portUint16, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error parsing port value: %s", err.Error())
	}

	switch proto {
	case "tcp":
		addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
		return addr, proto, int(portUint16), err
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
		return addr, proto, int(portUint16), err
	default:
		return nil, "", 0, fmt.Errorf("unknown protocol %s", proto)
	}
}
