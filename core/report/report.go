package report

import (
	"encoding/json"
	"fmt"
	"honeypot/core/transport/upstream"
)

type ReportResult struct {
	Typex    string `json:"type"`
	SourceIp string `json:"sourceIp"`
	Info     string `json:"info"`
}

func ReportToEdge(typex, sourceIp, info string) {
	reportResult := ReportResult{
		Typex:    typex,
		SourceIp: sourceIp,
		Info:     info,
	}
	fmt.Printf("publish msg\n")
	payload, err := json.Marshal(&reportResult)
	if err != nil {
		fmt.Errorf("json marshal err: %v\n", err)
	}
	upstream.Publish("HPReport", payload)
}
