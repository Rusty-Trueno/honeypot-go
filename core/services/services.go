// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package services

import (
	"context"
	"net"

	"github.com/BurntSushi/toml"
	"honeypot/core/event"
	"honeypot/core/pushers"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("services")

var (
	services = map[string]func(...ServicerFunc) Servicer{}
)

type ServicerFunc func(Servicer) error

func Register(key string, fn func(...ServicerFunc) Servicer) func(...ServicerFunc) Servicer {
	services[key] = fn
	return fn
}

func Range(fn func(string)) {
	for k := range services {
		fn(k)
	}
}

func Get(key string) (func(...ServicerFunc) Servicer, bool) {
	d := Dummy

	if fn, ok := services[key]; ok {
		return fn, true
	}

	return d, false
}

type CanHandlerer interface {
	CanHandle([]byte) bool
}

type Servicer interface {
	Handle(context.Context, net.Conn) error

	SetChannel(pushers.Channel)
}

func WithChannel(eb pushers.Channel) ServicerFunc {
	return func(d Servicer) error {
		d.SetChannel(eb)
		return nil
	}
}

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) ServicerFunc {
	return func(s Servicer) error {
		err := decoder.PrimitiveDecode(c, s)
		return err
	}
}

var (
	SensorLow = event.Sensor("services")

	EventOptions = event.NewWith(
		SensorLow,
	)
)
