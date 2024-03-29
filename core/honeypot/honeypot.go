package honeypot

import (
	"context"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/op/go-logging"
	"honeypot/core/pushers"
	bypass2 "honeypot/core/pushers/bypass"
	"honeypot/core/pushers/eventbus"
	"honeypot/core/pushers/timeseries"
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

	_ "honeypot/core/pushers/console"
	_ "honeypot/core/pushers/eventbus"
	_ "honeypot/core/pushers/timeseries"

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

var log = logging.MustGetLogger("honeypot")

type Honeypot struct {
	Name     string
	StopCh   chan bool
	ctx      context.Context
	cancel   context.CancelFunc
	Port     string
	Protocol string
	Switch   string
	Env      string
	l        net.Listener
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
						h.Switch = v
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
	h.stop()
}

func (h *Honeypot) stop() {
	if h.l != nil {
		err := h.l.Close()
		if err != nil {
			fmt.Errorf("close socker failed, err is %v", err)
		}
	}
	h.cancel()
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
					fmt.Printf("%s -> not listen\n", h.Port)
					status = "OFF"
				} else {
					fmt.Printf("%s -> listening\n", h.Port)
					status = "ON"
				}
			} else if h.Env == constant.Linux {
				if !linux.CheckPort(h.Port) {
					// 如果端口不再监听了
					fmt.Printf("%s -> not listen\n", h.Port)
					status = "OFF"
				} else {
					fmt.Printf("%s -> listening\n", h.Port)
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
	log.Infof("launching pot.......\n")
	// init event bus
	bc := pushers.NewBusChannel()
	bus := eventbus.New()
	bus.Subscribe(bc)
	bus.Subscribe(bypass2.Bk)
	bus.Subscribe(timeseries.Bk)
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

	addr, _, _, err := ToAddr(fmt.Sprintf("tcp/%s", h.Port))
	if err != nil {
		fmt.Errorf("Error parsing port string: %s", err.Error())
		return
	}
	if addr == nil {
		fmt.Errorf("Failed to bind: addr is nil")
		return
	}

	//start listen
	h.l, err = net.Listen(addr.Network(), addr.String())
	if err != nil {
		fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
		return
	}

	log.Infof("Listener started: tcp/%s", addr)

	ch := make(chan net.Conn)

	go func() {
		for {
			c, err := h.l.Accept()
			if err != nil {
				log.Errorf("Error accepting connection: %s", err.Error())
				break
			}

			ch <- c
		}
	}()

	incoming := make(chan net.Conn)
	var conn net.Conn
	go func() {
		for {
			select {
			case conn = <-ch:
				incoming <- conn
				// in case of goroutine starvation
				// with many connection and single procs
				runtime.Gosched()
			case <-h.ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-h.ctx.Done():
			return
		case conn := <-incoming:
			go service.Handle(h.ctx, conn)
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
