package timeseries

import (
	"encoding/json"
	"fmt"
	"honeypot/core/db/tdengine"
	"honeypot/core/event"
)

var (
	Bk Backend
)

func New() error {
	ch := make(chan map[string]interface{}, 100)
	Bk = Backend{
		ch: ch,
	}
	go Bk.run()
	return nil
}

type Backend struct {
	ch chan map[string]interface{}
}

func (b Backend) run() {
	for e := range b.ch {
		category, ok := e["category"]
		if !ok {
			fmt.Errorf("can't get category")
		}
		potName := fmt.Sprintf("%s", category)
		dataType, _ := json.Marshal(e)
		potData := string(dataType)
		err := tdengine.InsertPotData(potName, potData)
		if err != nil {
			fmt.Errorf("insert pot data to tdengine failed, error is %v", err)
		}
	}
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (b Backend) Send(e event.Event) {
	mp := make(map[string]interface{})

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	b.ch <- mp
}
