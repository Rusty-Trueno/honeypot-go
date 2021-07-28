package main

import (
	"flag"
	"honeypot/controller"
)

var (
	node = ""
)

func init() {
	flag.StringVar(&node, "node", "edge", "the node of this mapper")
}

func main() {
	controller.Run(node)
}
